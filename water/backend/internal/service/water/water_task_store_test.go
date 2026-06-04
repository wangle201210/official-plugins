// This file tests the watermark task status store.

package water

import (
	"context"
	"testing"
	"time"

	"lina-core/pkg/plugin/capability/contract"
)

// taskStoreCache records host cache writes for task-store tests.
type taskStoreCache struct {
	items         map[string]string
	lastNamespace string
	lastKey       string
	lastTTL       time.Duration
}

// newTaskStoreCache creates an empty task-store cache test double.
func newTaskStoreCache() *taskStoreCache {
	return &taskStoreCache{items: make(map[string]string)}
}

// Get returns one cached value.
func (c *taskStoreCache) Get(_ context.Context, namespace string, key string) (*contract.CacheItem, bool, error) {
	c.lastNamespace = namespace
	c.lastKey = key
	value, ok := c.items[namespace+"\x00"+key]
	if !ok {
		return nil, false, nil
	}
	return &contract.CacheItem{Key: key, ValueKind: contract.CacheValueKindString, Value: value}, true, nil
}

// Set records one cached value and TTL.
func (c *taskStoreCache) Set(_ context.Context, namespace string, key string, value string, ttl time.Duration) (*contract.CacheItem, error) {
	c.lastNamespace = namespace
	c.lastKey = key
	c.lastTTL = ttl
	c.items[namespace+"\x00"+key] = value
	return &contract.CacheItem{Key: key, ValueKind: contract.CacheValueKindString, Value: value}, nil
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
