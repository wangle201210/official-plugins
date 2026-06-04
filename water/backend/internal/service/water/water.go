// Package water implements watermark snapshot processing for the water source plugin.
package water

import (
	"context"
	"time"

	"github.com/gogf/gf/v2/errors/gerror"

	"lina-core/pkg/plugin/capability/contract"
	mediastrategy "lina-plugin-media/backend/provider/strategy"
)

// Service defines the water plugin service contract.
type Service interface {
	// SubmitSnap submits one asynchronous watermark snapshot task.
	SubmitSnap(ctx context.Context, in SubmitSnapInput) (*SubmitSnapOutput, error)
	// Preview synchronously renders one watermark preview.
	Preview(ctx context.Context, in PreviewInput) (*ProcessOutput, error)
	// GetTask returns one recent watermark task snapshot.
	GetTask(ctx context.Context, taskID string) (*TaskSnapshot, error)
}

// Interface compliance assertion for the default water service implementation.
var _ Service = (*serviceImpl)(nil)

// serviceImpl implements Service.
type serviceImpl struct {
	queue            *taskQueue             // queue executes asynchronous watermark tasks.
	store            *taskStore             // store keeps recent task status snapshots in host cache.
	strategyResolver mediastrategy.Resolver // strategyResolver resolves media-owned strategy bindings.
}

// taskCache defines the host cache operations water uses for task snapshots.
type taskCache interface {
	// Get returns the cached task snapshot payload for key when it exists.
	Get(ctx context.Context, namespace string, key string) (*contract.CacheItem, bool, error)
	// Set stores the task snapshot payload for key with a finite TTL.
	Set(ctx context.Context, namespace string, key string, value string, ttl time.Duration) (*contract.CacheItem, error)
}

// New creates and returns the shared water service instance.
func New(cacheSvc contract.CacheService, strategyResolver mediastrategy.Resolver) (Service, error) {
	if cacheSvc == nil {
		return nil, gerror.New("water service requires host cache service")
	}
	if strategyResolver == nil {
		return nil, gerror.New("water service requires media strategy resolver")
	}
	store := newTaskStore(cacheSvc)
	service := &serviceImpl{
		store:            store,
		strategyResolver: strategyResolver,
	}
	service.queue = newTaskQueue(store, service.processSnapshot)
	return service, nil
}
