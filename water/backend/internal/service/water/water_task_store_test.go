// This file tests the watermark task status store.

package water

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"testing"
	"time"

	"lina-core/pkg/plugin/capability/cachecap"
)

// taskStoreCache records host cache writes for task-store tests.
type taskStoreCache struct {
	items         map[string]string
	lastNamespace string
	lastKey       string
	lastTTL       time.Duration
	maxValueBytes int
}

// newTaskStoreCache creates an empty task-store cache test double.
func newTaskStoreCache() *taskStoreCache {
	return &taskStoreCache{items: make(map[string]string)}
}

// Get returns one cached value.
func (c *taskStoreCache) Get(_ context.Context, namespace string, key string) (*cachecap.CacheItem, bool, error) {
	c.lastNamespace = namespace
	c.lastKey = key
	value, ok := c.items[namespace+"\x00"+key]
	if !ok {
		return nil, false, nil
	}
	return &cachecap.CacheItem{Key: key, ValueKind: cachecap.CacheValueKindString, Value: value}, true, nil
}

// Set records one cached value and TTL.
func (c *taskStoreCache) Set(_ context.Context, namespace string, key string, value string, ttl time.Duration) (*cachecap.CacheItem, error) {
	if c.maxValueBytes > 0 && len(value) > c.maxValueBytes {
		return nil, fmt.Errorf("cache value exceeds %d bytes", c.maxValueBytes)
	}
	c.lastNamespace = namespace
	c.lastKey = key
	c.lastTTL = ttl
	c.items[namespace+"\x00"+key] = value
	return &cachecap.CacheItem{Key: key, ValueKind: cachecap.CacheValueKindString, Value: value}, nil
}

// TestTaskStoreUsesHostCache verifies task snapshots are stored in host cache.
func TestTaskStoreUsesHostCache(t *testing.T) {
	ctx := context.Background()
	cacheSvc := newTaskStoreCache()
	store := newTaskStore(cacheSvc)
	if err := store.create(ctx, "task-1", SubmitSnapInput{Tenant: "tenant-a", DeviceId: "device-a"}); err != nil {
		t.Fatalf("create task snapshot: %v", err)
	}
	if cacheSvc.lastNamespace != "task-status" {
		t.Fatalf("expected task-status namespace, got %s", cacheSvc.lastNamespace)
	}
	if cacheSvc.lastKey != "water:task:task-1" {
		t.Fatalf("expected water task cache key, got %s", cacheSvc.lastKey)
	}
	if cacheSvc.lastTTL != 12*time.Hour {
		t.Fatalf("expected 12h TTL, got %s", cacheSvc.lastTTL)
	}
	if err := store.update(ctx, "task-1", func(record *taskRecord) {
		record.Status = TaskStatusSuccess
		record.Message = "完成"
		record.Success = true
	}); err != nil {
		t.Fatalf("update task snapshot: %v", err)
	}
	task, err := store.get(ctx, "task-1")
	if err != nil {
		t.Fatalf("expected task-1 to exist: %v", err)
	}
	if task.Tenant != "tenant-a" || task.Status != TaskStatusSuccess || !task.Success {
		t.Fatalf("unexpected task snapshot: %+v", task)
	}
}

// TestTaskStoreUpdateStripsLegacyCachedImage verifies updates can recover
// status records created by older code that cached a large image payload.
func TestTaskStoreUpdateStripsLegacyCachedImage(t *testing.T) {
	ctx := context.Background()
	cacheSvc := newTaskStoreCache()
	cacheSvc.maxValueBytes = 4096
	store := newTaskStore(cacheSvc)
	legacyRecord := &taskRecord{
		TaskSnapshot: TaskSnapshot{
			TaskId:    "task-legacy-image",
			Status:    TaskStatusProcessing,
			Message:   "处理中",
			Tenant:    "tenant-a",
			DeviceId:  "device-a",
			Image:     "data:image/png;base64," + strings.Repeat("a", 8192),
			CreatedAt: time.Now().UnixMilli(),
			UpdatedAt: time.Now().UnixMilli(),
		},
	}
	payload, err := json.Marshal(legacyRecord)
	if err != nil {
		t.Fatalf("marshal legacy task snapshot: %v", err)
	}
	cacheKey := taskStatusCacheNamespace + "\x00" + taskStatusCacheKey("task-legacy-image")
	cacheSvc.items[cacheKey] = string(payload)

	if err = store.update(ctx, "task-legacy-image", func(record *taskRecord) {
		record.Status = TaskStatusSuccess
		record.Success = true
		record.Message = "处理完成"
		record.Source = StrategySourceGlobal
		record.SourceLabel = strategySourceLabel(StrategySourceGlobal)
	}); err != nil {
		t.Fatalf("update legacy task snapshot: %v", err)
	}
	task, err := store.get(ctx, "task-legacy-image")
	if err != nil {
		t.Fatalf("get task snapshot: %v", err)
	}
	if task.Image != "" {
		t.Fatalf("expected cached image to be stripped, got %d bytes", len(task.Image))
	}
	if len(cacheSvc.items[cacheKey]) > 4096 {
		t.Fatalf("expected compact cached task status, got %d bytes", len(cacheSvc.items[cacheKey]))
	}
}
