// cron_test.go verifies scheduled-job registration and primary-node gating.
package cron

import (
	"context"
	"testing"
	"time"

	"lina-core/pkg/plugin/capability"
	"lina-core/pkg/plugin/pluginhost"
	feedingsvc "lina-plugin-sicau-niu/backend/internal/service/feeding"
	irontransportsvc "lina-plugin-sicau-niu/backend/internal/service/irontransport"
)

type fakeJobsRegistrar struct {
	primary  bool
	names    []string
	handlers map[string]pluginhost.JobHandler
}

func (r *fakeJobsRegistrar) Add(ctx context.Context, pattern, name string, handler pluginhost.JobHandler) error {
	return r.AddWithMetadata(ctx, pattern, name, name, "", handler)
}

func (r *fakeJobsRegistrar) AddWithMetadata(_ context.Context, _, name, _, _ string, handler pluginhost.JobHandler) error {
	r.names = append(r.names, name)
	if r.handlers == nil {
		r.handlers = make(map[string]pluginhost.JobHandler)
	}
	r.handlers[name] = handler
	return nil
}

func (r *fakeJobsRegistrar) IsPrimaryNode() bool           { return r.primary }
func (r *fakeJobsRegistrar) Services() capability.Services { return nil }

type fakeTransportService struct {
	irontransportsvc.Service
	called bool
}

func (s *fakeTransportService) ExpireInactive(_ context.Context, _ time.Time) (int, error) {
	s.called = true
	return 1, nil
}

type fakeIronRefresher struct {
	called bool
}

func (r *fakeIronRefresher) Refresh(_ context.Context) (*feedingsvc.IronLocationRefreshResult, error) {
	r.called = true
	return &feedingsvc.IronLocationRefreshResult{}, nil
}

func TestRegisterAlwaysAddsExpiryAndConditionallyAddsIronRefresh(t *testing.T) {
	transport := &fakeTransportService{}
	refresher := &fakeIronRefresher{}
	service, err := New(transport, refresher, Config{IronLocationEnabled: true, IronLocationInterval: time.Minute})
	if err != nil {
		t.Fatalf("New returned error: %v", err)
	}
	registrar := &fakeJobsRegistrar{primary: true}
	if err = service.Register(context.Background(), registrar); err != nil {
		t.Fatalf("Register returned error: %v", err)
	}
	if len(registrar.names) != 2 || registrar.names[0] != transportExpiryName || registrar.names[1] != ironLocationRefreshName {
		t.Fatalf("unexpected registered jobs: %#v", registrar.names)
	}
	if err = registrar.handlers[transportExpiryName](context.Background()); err != nil {
		t.Fatalf("expiry handler returned error: %v", err)
	}
	if err = registrar.handlers[ironLocationRefreshName](context.Background()); err != nil {
		t.Fatalf("refresh handler returned error: %v", err)
	}
	if !transport.called || !refresher.called {
		t.Fatalf("expected both primary-node handlers to run: transport=%v refresher=%v", transport.called, refresher.called)
	}
}

func TestRegisteredJobsSkipNonPrimaryNode(t *testing.T) {
	transport := &fakeTransportService{}
	service, err := New(transport, nil, Config{})
	if err != nil {
		t.Fatalf("New returned error: %v", err)
	}
	registrar := &fakeJobsRegistrar{}
	if err = service.Register(context.Background(), registrar); err != nil {
		t.Fatalf("Register returned error: %v", err)
	}
	if len(registrar.names) != 1 {
		t.Fatalf("expected only expiry job, got %#v", registrar.names)
	}
	if err = registrar.handlers[transportExpiryName](context.Background()); err != nil {
		t.Fatalf("non-primary handler returned error: %v", err)
	}
	if transport.called {
		t.Fatal("expected non-primary node to skip expiry")
	}
}

func TestNewRejectsMissingDependencies(t *testing.T) {
	if _, err := New(nil, nil, Config{}); err == nil {
		t.Fatal("expected missing transport service to fail")
	}
	if _, err := New(&fakeTransportService{}, nil, Config{IronLocationEnabled: true, IronLocationInterval: time.Minute}); err == nil {
		t.Fatal("expected enabled refresh without refresher to fail")
	}
}
