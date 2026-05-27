// This file verifies linapro-monitor-server plugin callback wiring helpers.

package backend

import (
	"context"
	"testing"
	"time"

	pluginconfig "lina-core/pkg/plugin/capability/config"
	plugincontract "lina-core/pkg/plugin/capability/contract"
	"lina-core/pkg/plugin/capability/orgcap"
	"lina-core/pkg/plugin/capability/tenantcap"
	monitorsvc "lina-plugin-linapro-monitor-server/backend/internal/service/monitor"
)

// fakeCapabilities publishes only the config dependency used by this test file.
type fakeCapabilities struct {
	configSvc plugincontract.ConfigService
}

// APIDoc returns no apidoc service.
func (s fakeCapabilities) APIDoc() plugincontract.APIDocService { return nil }

// Auth returns no auth service.
func (s fakeCapabilities) Auth() plugincontract.AuthService { return nil }

// BizCtx returns no bizctx service.
func (s fakeCapabilities) BizCtx() plugincontract.BizCtxService { return nil }

// Cache returns no cache service.
func (s fakeCapabilities) Cache() plugincontract.CacheService { return nil }

// Config returns the configured fake config service.
func (s fakeCapabilities) Config() plugincontract.ConfigService { return s.configSvc }

// HostConfig returns no host config service.
func (s fakeCapabilities) HostConfig() plugincontract.HostConfigService { return nil }

// I18n returns no i18n service.
func (s fakeCapabilities) I18n() plugincontract.I18nService { return nil }

// Manifest returns no manifest service.
func (s fakeCapabilities) Manifest() plugincontract.ManifestService { return nil }

// Notify returns no notify service.
func (s fakeCapabilities) Notify() plugincontract.NotifyService { return nil }

// Org returns the default organization capability fallback service.
func (s fakeCapabilities) Org() orgcap.Service { return orgcap.New(nil) }

// PluginLifecycle returns no plugin lifecycle service.
func (s fakeCapabilities) PluginLifecycle() plugincontract.PluginLifecycleService { return nil }

// PluginState returns no plugin-state service.
func (s fakeCapabilities) PluginState() plugincontract.PluginStateService { return nil }

// Route returns no route service.
func (s fakeCapabilities) Route() plugincontract.RouteService { return nil }

// Session returns no session service.
func (s fakeCapabilities) Session() plugincontract.SessionService { return nil }

// TenantFilter returns no tenant-filter service.
func (s fakeCapabilities) TenantFilter() plugincontract.TenantFilterService { return nil }

// Tenant returns the default tenant capability fallback service.
func (s fakeCapabilities) Tenant() tenantcap.Service { return tenantcap.New(nil, nil) }

// fakeMonitorService records callback usage without touching the database or host metrics.
type fakeMonitorService struct {
	// collected reports whether CollectAndStore was called.
	collected bool
	// cleanupCalled reports whether CleanupStale was called.
	cleanupCalled bool
	// cleanupThreshold records the threshold passed to CleanupStale.
	cleanupThreshold time.Duration
}

// CollectAndStore records one collection callback.
func (s *fakeMonitorService) CollectAndStore(ctx context.Context) {
	s.collected = true
}

// Collect satisfies monitorsvc.Service for tests.
func (s *fakeMonitorService) Collect(ctx context.Context) *monitorsvc.MonitorData {
	return nil
}

// GetDBInfo satisfies monitorsvc.Service for tests.
func (s *fakeMonitorService) GetDBInfo(ctx context.Context) *monitorsvc.DBInfo {
	return nil
}

// GetLatest satisfies monitorsvc.Service for tests.
func (s *fakeMonitorService) GetLatest(ctx context.Context, nodeName string) ([]*monitorsvc.NodeMonitorData, error) {
	return nil, nil
}

// CleanupStale records one cleanup callback.
func (s *fakeMonitorService) CleanupStale(ctx context.Context, threshold time.Duration) (int64, error) {
	s.cleanupCalled = true
	s.cleanupThreshold = threshold
	return 0, nil
}

// TestCollectSnapshotUsesInjectedService verifies cron collection reuses the provided service instance.
func TestCollectSnapshotUsesInjectedService(t *testing.T) {
	monitorSvc := &fakeMonitorService{}

	if err := collectSnapshot(context.Background(), monitorSvc); err != nil {
		t.Fatalf("collect snapshot: %v", err)
	}

	if !monitorSvc.collected {
		t.Fatal("expected injected monitor service to collect")
	}
}

// TestCleanupSnapshotsSkipsNonPrimaryNode verifies cleanup is skipped outside the primary node.
func TestCleanupSnapshotsSkipsNonPrimaryNode(t *testing.T) {
	monitorSvc := &fakeMonitorService{}

	if err := cleanupSnapshots(context.Background(), false, nil, monitorSvc); err != nil {
		t.Fatalf("cleanup snapshots: %v", err)
	}

	if monitorSvc.cleanupCalled {
		t.Fatal("expected non-primary node to skip cleanup")
	}
}

// TestCleanupSnapshotsUsesInjectedServiceOnPrimaryNode verifies cleanup uses the shared service instance.
func TestCleanupSnapshotsUsesInjectedServiceOnPrimaryNode(t *testing.T) {
	configSvc := newPluginTestConfigService(t, `
monitor:
  interval: 30s
  retentionMultiplier: 4
`)

	monitorSvc := &fakeMonitorService{}
	capabilities := fakeCapabilities{
		configSvc: configSvc,
	}

	if err := cleanupSnapshots(context.Background(), true, capabilities, monitorSvc); err != nil {
		t.Fatalf("cleanup snapshots: %v", err)
	}

	if !monitorSvc.cleanupCalled {
		t.Fatal("expected injected monitor service to clean up")
	}
	if monitorSvc.cleanupThreshold != 2*time.Minute {
		t.Fatalf("expected cleanup threshold 2m, got %s", monitorSvc.cleanupThreshold)
	}
}

// newPluginTestConfigService builds a scoped plugin config service from artifact content.
func newPluginTestConfigService(t *testing.T, content string) plugincontract.ConfigService {
	t.Helper()

	return pluginconfig.NewFactory(t.TempDir(), t.TempDir()).
		WithArtifactConfig("linapro-monitor-server", []byte(content)).
		ForPlugin("linapro-monitor-server")
}
