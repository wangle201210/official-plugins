// This file tests asynchronous watermark local queue state transitions.

package water

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

// TestLocalTaskQueueSubmitStaysProcessLocal verifies submit only delivers to
// the in-process channel and does not persist task payloads to host cache or Redis.
func TestLocalTaskQueueSubmitStaysProcessLocal(t *testing.T) {
	ctx := context.Background()
	queue := newLocalTaskQueue(nil, nil, defaultConsumerCount)
	queue.startOnce.Do(func() {})

	task := &watermarkTask{
		id:      "task-local-only",
		ctx:     ctx,
		request: SubmitSnapInput{Tenant: "tenant-a", DeviceId: "device-a"},
	}
	if err := queue.submit(ctx, task); err != nil {
		t.Fatalf("submit local task: %v", err)
	}

	select {
	case got := <-queue.tasks:
		if got != task {
			t.Fatalf("expected queued task pointer %p, got %p", task, got)
		}
	default:
		t.Fatal("expected task to be queued in local channel")
	}
}

// TestTaskQueueDoesNotCacheLargeImageResult verifies async status snapshots
// stay below the host cache value limit while callbacks still receive images.
func TestTaskQueueDoesNotCacheLargeImageResult(t *testing.T) {
	ctx := context.Background()
	image := "data:image/png;base64," + strings.Repeat("a", 8192)
	received := make(chan snapPayload, 1)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := r.Body.Close(); err != nil {
				t.Errorf("close request body: %v", err)
			}
		}()
		var payload snapPayload
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			t.Errorf("decode callback payload: %v", err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		received <- payload
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	cacheSvc := newTaskStoreCache()
	cacheSvc.maxValueBytes = 4096
	store := newTaskStore(cacheSvc)
	if err := store.create(ctx, "task-large-image", SubmitSnapInput{Tenant: "tenant-a", DeviceId: "device-a"}); err != nil {
		t.Fatalf("create task snapshot: %v", err)
	}

	queue := newLocalTaskQueue(store, func(context.Context, SubmitSnapInput) (*ProcessOutput, error) {
		return &ProcessOutput{
			Success:     true,
			Status:      TaskStatusSuccess,
			Message:     "处理完成",
			Image:       image,
			Source:      StrategySourceGlobal,
			SourceLabel: strategySourceLabel(StrategySourceGlobal),
		}, nil
	}, defaultConsumerCount)
	queue.processTask(1, &watermarkTask{
		id:  "task-large-image",
		ctx: ctx,
		request: SubmitSnapInput{
			CallbackUrl: server.URL,
			Tenant:      "tenant-a",
			DeviceId:    "device-a",
		},
	})

	task, err := store.get(ctx, "task-large-image")
	if err != nil {
		t.Fatalf("get task snapshot: %v", err)
	}
	if task.Status != TaskStatusSuccess || !task.Success {
		t.Fatalf("expected successful task status, got %+v", task)
	}
	if task.Image != "" {
		t.Fatalf("expected async task status image to stay empty, got %d bytes", len(task.Image))
	}
	cacheValue := cacheSvc.items[taskStatusCacheNamespace+"\x00"+taskStatusCacheKey("task-large-image")]
	if len(cacheValue) > 4096 {
		t.Fatalf("expected cached task status to stay within 4096 bytes, got %d", len(cacheValue))
	}

	select {
	case payload := <-received:
		if payload.Image != image {
			t.Fatalf("callback image length got %d want %d", len(payload.Image), len(image))
		}
	case <-time.After(time.Second):
		t.Fatal("expected callback payload")
	}
}
