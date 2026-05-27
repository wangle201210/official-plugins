// This file implements asynchronous watermark task execution.

package water

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"sync"
	"time"

	"github.com/gogf/gf/v2/frame/g"

	"lina-core/pkg/bizerr"
	"lina-core/pkg/logger"
)

// watermarkTask is one queued asynchronous snapshot task.
type watermarkTask struct {
	id      string
	ctx     context.Context
	request SubmitSnapInput
}

// taskQueue owns worker startup and queued task delivery.
type taskQueue struct {
	store     *taskStore
	processor func(ctx context.Context, in SubmitSnapInput) (*ProcessOutput, error)
	tasks     chan *watermarkTask
	startOnce sync.Once
}

// newTaskQueue creates one task queue with lazy workers.
func newTaskQueue(store *taskStore, processor func(ctx context.Context, in SubmitSnapInput) (*ProcessOutput, error)) *taskQueue {
	return &taskQueue{
		store:     store,
		processor: processor,
		tasks:     make(chan *watermarkTask, defaultTaskQueueCapacity),
	}
}

// submit enqueues one asynchronous task.
func (q *taskQueue) submit(ctx context.Context, task *watermarkTask) error {
	q.start(ctx)
	select {
	case q.tasks <- task:
		return nil
	default:
		return bizerr.NewCode(CodeWaterTaskQueueFull)
	}
}

// start lazily starts workers.
func (q *taskQueue) start(ctx context.Context) {
	q.startOnce.Do(func() {
		consumerCount := g.Cfg().MustGet(ctx, "water.consumerCount", defaultConsumerCount).Int()
		if consumerCount < 1 {
			consumerCount = defaultConsumerCount
		}
		if consumerCount > maxConsumerCount {
			consumerCount = maxConsumerCount
		}
		logger.Infof(ctx, "启动 %d 个水印消费者", consumerCount)
		for i := 0; i < consumerCount; i++ {
			go q.consume(ctx, i+1)
		}
	})
}

// consume processes queued tasks until the process exits.
func (q *taskQueue) consume(rootCtx context.Context, consumerID int) {
	logger.Infof(rootCtx, "水印消费者 %d 已启动", consumerID)
	for task := range q.tasks {
		if task == nil {
			continue
		}
		q.processTask(consumerID, task)
	}
}

// processTask executes one asynchronous task and records the final status.
func (q *taskQueue) processTask(consumerID int, task *watermarkTask) {
	start := time.Now()
	ctx := task.ctx
	if q == nil || q.processor == nil {
		logger.Errorf(ctx, "消费者 %d: 水印处理器未初始化 %s", consumerID, task.id)
		if updateErr := q.store.update(ctx, task.id, func(record *taskRecord) {
			record.Status = TaskStatusFailed
			record.Success = false
			record.Message = "处理失败"
			record.Error = "water processor is not initialized"
			record.DurationMs = time.Since(start).Milliseconds()
		}); updateErr != nil {
			logger.Errorf(ctx, "消费者 %d: 更新水印任务失败状态失败 %s: %v", consumerID, task.id, updateErr)
		}
		return
	}
	logger.Infof(ctx, "消费者 %d 开始处理水印任务: %s", consumerID, task.id)
	if err := q.store.update(ctx, task.id, func(record *taskRecord) {
		record.Status = TaskStatusProcessing
		record.Message = "处理中"
	}); err != nil {
		logger.Errorf(ctx, "消费者 %d: 更新水印任务处理中状态失败 %s: %v", consumerID, task.id, err)
	}

	output, err := q.processor(ctx, task.request)
	if err != nil {
		logger.Errorf(ctx, "消费者 %d: 水印任务失败 %s: %v", consumerID, task.id, err)
		if updateErr := q.store.update(ctx, task.id, func(record *taskRecord) {
			record.Status = TaskStatusFailed
			record.Success = false
			record.Message = "处理失败"
			record.Error = err.Error()
			record.DurationMs = time.Since(start).Milliseconds()
		}); updateErr != nil {
			logger.Errorf(ctx, "消费者 %d: 更新水印任务失败状态失败 %s: %v", consumerID, task.id, updateErr)
		}
		return
	}

	if err = q.store.update(ctx, task.id, func(record *taskRecord) {
		record.Status = output.Status
		record.Success = output.Success
		record.Message = output.Message
		record.Error = output.Error
		record.Image = output.Image
		record.StrategyId = output.StrategyId
		record.StrategyName = output.StrategyName
		record.Source = output.Source
		record.SourceLabel = output.SourceLabel
		record.DurationMs = time.Since(start).Milliseconds()
	}); err != nil {
		logger.Errorf(ctx, "消费者 %d: 更新水印任务完成状态失败 %s: %v", consumerID, task.id, err)
	}

	callbackURL := normalizedCallbackURL(task.request)
	if callbackURL != "" {
		payload := buildCallbackPayload(task.request, output.Image)
		if err := sendResultToURL(ctx, callbackURL, payload); err != nil {
			logger.Errorf(ctx, "消费者 %d: 发送水印回调失败 %s: %v", consumerID, task.id, err)
			if updateErr := q.store.update(ctx, task.id, func(record *taskRecord) {
				record.Error = err.Error()
			}); updateErr != nil {
				logger.Errorf(ctx, "消费者 %d: 更新水印任务回调错误失败 %s: %v", consumerID, task.id, updateErr)
			}
		}
	}
	logger.Infof(ctx, "消费者 %d 完成水印任务: %s, 耗时: %s", consumerID, task.id, time.Since(start))
}

// generateTaskID creates a unique task identifier.
func generateTaskID() string {
	timestamp := time.Now().UnixNano()
	randomBytes := make([]byte, 8)
	if _, err := rand.Read(randomBytes); err != nil {
		return fmt.Sprintf("wm_%d", timestamp)
	}
	randomStr := base64.RawURLEncoding.EncodeToString(randomBytes)
	if len(randomStr) > 8 {
		randomStr = randomStr[:8]
	}
	return fmt.Sprintf("wm_%d_%s", timestamp, randomStr)
}
