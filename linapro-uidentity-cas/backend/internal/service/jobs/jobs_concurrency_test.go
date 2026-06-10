// This file verifies the scheduled-job concurrency guard: non-primary nodes
// skip execution and a busy job is not overlapped by the next trigger.

package jobs

import (
	"context"
	"sync"
	"testing"
)

// TestGuardConcurrencySkipsNonPrimaryNode verifies a non-primary node never
// runs the handler, replicating the old single-owner execution.
func TestGuardConcurrencySkipsNonPrimaryNode(t *testing.T) {
	t.Parallel()

	s := &serviceImpl{}
	ran := false
	guarded := s.guardConcurrency(&fakeCronRegistrar{notPrimary: true}, "job", func(context.Context) error {
		ran = true
		return nil
	})
	if err := guarded(context.Background()); err != nil {
		t.Fatalf("guarded handler returned error: %v", err)
	}
	if ran {
		t.Fatal("handler must not run on a non-primary node")
	}
}

// TestGuardConcurrencySkipsOverlappingRun verifies a second trigger is skipped
// while the first run still holds the per-job lock.
func TestGuardConcurrencySkipsOverlappingRun(t *testing.T) {
	t.Parallel()

	s := &serviceImpl{}
	registrar := &fakeCronRegistrar{}
	started := make(chan struct{})
	release := make(chan struct{})
	var runs int
	var mu sync.Mutex

	guarded := s.guardConcurrency(registrar, "job", func(context.Context) error {
		mu.Lock()
		runs++
		first := runs == 1
		mu.Unlock()
		if first {
			// Only the first run holds the lock long enough to test overlap.
			close(started)
			<-release
		}
		return nil
	})

	done := make(chan struct{})
	go func() {
		_ = guarded(context.Background())
		close(done)
	}()
	<-started

	// A second trigger while the first run is in progress must be skipped.
	if err := guarded(context.Background()); err != nil {
		t.Fatalf("overlapping guarded handler returned error: %v", err)
	}
	mu.Lock()
	if runs != 1 {
		mu.Unlock()
		t.Fatalf("expected overlapping run to be skipped, runs=%d", runs)
	}
	mu.Unlock()

	close(release)
	<-done

	// After the first run releases, the job can run again.
	if err := guarded(context.Background()); err != nil {
		t.Fatalf("post-release guarded handler returned error: %v", err)
	}
	mu.Lock()
	defer mu.Unlock()
	if runs != 2 {
		t.Fatalf("expected job to run again after release, runs=%d", runs)
	}
}
