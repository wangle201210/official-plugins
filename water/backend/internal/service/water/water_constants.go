// This file defines water service constants and enum-like values.

package water

import "time"

// TaskStatus describes the lifecycle state of a watermark task.
type TaskStatus string

// StrategySource describes the source from which a media strategy was resolved.
type StrategySource string

const (
	// TaskStatusQueued means the task is waiting for a worker.
	TaskStatusQueued TaskStatus = "queued"
	// TaskStatusProcessing means the task is currently being processed.
	TaskStatusProcessing TaskStatus = "processing"
	// TaskStatusSuccess means the task produced a watermarked image.
	TaskStatusSuccess TaskStatus = "success"
	// TaskStatusSkipped means no matching strategy or watermark config was found and the original image was returned.
	TaskStatusSkipped TaskStatus = "skipped"
	// TaskStatusFailed means the task failed.
	TaskStatusFailed TaskStatus = "failed"
)

const (
	// StrategySourceTenantDevice means the tenant-device binding matched.
	StrategySourceTenantDevice StrategySource = "tenantDevice"
	// StrategySourceDevice means the device binding matched.
	StrategySourceDevice StrategySource = "device"
	// StrategySourceTenant means the tenant binding matched.
	StrategySourceTenant StrategySource = "tenant"
	// StrategySourceGlobal means the global strategy matched.
	StrategySourceGlobal StrategySource = "global"
	// StrategySourceNone means no enabled media strategy matched.
	StrategySourceNone StrategySource = "none"
)

const (
	// defaultTaskQueueCapacity caps queued asynchronous tasks.
	defaultTaskQueueCapacity = 1024
	// defaultConsumerCount is used when water.consumerCount is not configured.
	defaultConsumerCount = 1
	// maxConsumerCount protects the host from accidental excessive worker creation.
	maxConsumerCount = 32
	// defaultCallbackTimeout is the callback HTTP client timeout.
	defaultCallbackTimeout = 30 * time.Second
	// defaultStrategyResolverTimeout is the remote media strategy API timeout.
	defaultStrategyResolverTimeout = 10 * time.Second
	// defaultTaskStatusTTL limits how long asynchronous task snapshots remain queryable.
	defaultTaskStatusTTL = 12 * time.Hour
	// defaultFontSize is used when a strategy omits fontSize.
	defaultFontSize = 32
	// defaultWatermarkOpacity is used when a strategy omits opacity.
	defaultWatermarkOpacity = 0.15
	// taskStatusCacheNamespace scopes water task snapshots inside the plugin cache.
	taskStatusCacheNamespace = "task-status"
	// taskStatusCacheKeyPrefix keeps task keys readable in the host cache.
	taskStatusCacheKeyPrefix = "water:task:"
	// strategyResolverPath is the mediaopen endpoint used by remote water clusters.
	strategyResolverPath = "/api/v1/strategies/resolve"
	// strategyResolverAPIKeyHeader is the mediaopen internal API key header.
	strategyResolverAPIKeyHeader = "X-Inner-Api-Key"
	// maxStrategyResolverResponseBytes bounds remote media strategy API response decoding.
	maxStrategyResolverResponseBytes = 1 << 20
)
