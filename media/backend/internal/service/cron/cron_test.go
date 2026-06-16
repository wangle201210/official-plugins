// This file verifies media cron startup error handling.

package cron

import (
	"context"
	"testing"

	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/os/gcron"
)

// TestStartReturnsCronRegistrationError verifies startup callers see cron registration failures.
func TestStartReturnsCronRegistrationError(t *testing.T) {
	original := addSingletonCron
	t.Cleanup(func() {
		addSingletonCron = original
	})

	expected := gerror.New("cron registration failed")
	addSingletonCron = func(ctx context.Context, pattern string, job gcron.JobFunc, name ...string) (*gcron.Entry, error) {
		return nil, expected
	}

	err := (&serviceImpl{}).Start(context.Background())
	if err != expected {
		t.Fatalf("expected cron registration error to be returned, got %v", err)
	}
}

// TestStartRetriesAfterRegistrationError verifies failed registration does not mark the service started.
func TestStartRetriesAfterRegistrationError(t *testing.T) {
	original := addSingletonCron
	t.Cleanup(func() {
		addSingletonCron = original
	})

	calls := 0
	addSingletonCron = func(ctx context.Context, pattern string, job gcron.JobFunc, name ...string) (*gcron.Entry, error) {
		calls++
		if calls == 1 {
			return nil, gerror.New("first failure")
		}
		return &gcron.Entry{}, nil
	}

	svc := &serviceImpl{}
	if err := svc.Start(context.Background()); err == nil {
		t.Fatal("expected first start to fail")
	}
	if err := svc.Start(context.Background()); err != nil {
		t.Fatalf("expected second start to retry and succeed: %v", err)
	}
	if calls != 2 {
		t.Fatalf("expected two registration attempts, got %d", calls)
	}
}
