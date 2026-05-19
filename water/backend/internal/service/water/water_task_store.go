// This file implements the host-cache-backed watermark task status store.

package water

import (
	"context"
	"encoding/json"

	"github.com/gogf/gf/v2/os/gtime"

	"lina-core/pkg/bizerr"
)

// taskRecord stores one mutable task status snapshot.
type taskRecord struct {
	TaskSnapshot
}

// taskStore keeps recent task snapshots in the host plugin cache.
type taskStore struct {
	cache taskCache
}

// newTaskStore creates one host-cache-backed task status store.
func newTaskStore(cache taskCache) *taskStore {
	return &taskStore{cache: cache}
}

// create inserts one queued task snapshot.
func (s *taskStore) create(ctx context.Context, taskID string, in SubmitSnapInput) error {
	now := gtime.Now().String()
	record := &taskRecord{
		TaskSnapshot: TaskSnapshot{
			TaskId:      taskID,
			Status:      TaskStatusQueued,
			Message:     "等待处理",
			Tenant:      in.Tenant,
			DeviceId:    in.DeviceId,
			Source:      StrategySourceNone,
			SourceLabel: strategySourceLabel(StrategySourceNone),
			CreatedAt:   now,
			UpdatedAt:   now,
		},
	}
	return s.save(ctx, record)
}

// update mutates one task snapshot if it still exists.
func (s *taskStore) update(ctx context.Context, taskID string, mutate func(record *taskRecord)) error {
	record, err := s.record(ctx, taskID)
	if err != nil {
		return err
	}
	mutate(record)
	record.UpdatedAt = gtime.Now().String()
	return s.save(ctx, record)
}

// get returns one task snapshot by ID.
func (s *taskStore) get(ctx context.Context, taskID string) (*TaskSnapshot, error) {
	record, err := s.record(ctx, taskID)
	if err != nil {
		return nil, err
	}
	snapshot := record.TaskSnapshot
	return &snapshot, nil
}

// record loads one mutable task record from host cache.
func (s *taskStore) record(ctx context.Context, taskID string) (*taskRecord, error) {
	if s == nil || s.cache == nil {
		return nil, bizerr.NewCode(CodeWaterTaskCacheFailed)
	}
	item, found, err := s.cache.Get(ctx, taskStatusCacheNamespace, taskStatusCacheKey(taskID))
	if err != nil {
		return nil, bizerr.WrapCode(err, CodeWaterTaskCacheFailed)
	}
	if !found || item == nil {
		return nil, bizerr.NewCode(CodeWaterTaskNotFound)
	}
	var record taskRecord
	if err = json.Unmarshal([]byte(item.Value), &record); err != nil {
		return nil, bizerr.WrapCode(err, CodeWaterTaskCacheFailed)
	}
	return &record, nil
}

// save stores one task record in host cache with the configured retention TTL.
func (s *taskStore) save(ctx context.Context, record *taskRecord) error {
	if s == nil || s.cache == nil || record == nil {
		return bizerr.NewCode(CodeWaterTaskCacheFailed)
	}
	payload, err := json.Marshal(record)
	if err != nil {
		return bizerr.WrapCode(err, CodeWaterTaskCacheFailed)
	}
	if _, err = s.cache.Set(ctx, taskStatusCacheNamespace, taskStatusCacheKey(record.TaskId), string(payload), defaultTaskStatusTTL); err != nil {
		return bizerr.WrapCode(err, CodeWaterTaskCacheFailed)
	}
	return nil
}

// taskStatusCacheKey builds the host cache key for one water task snapshot.
func taskStatusCacheKey(taskID string) string {
	return taskStatusCacheKeyPrefix + taskID
}
