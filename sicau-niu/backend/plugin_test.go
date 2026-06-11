// plugin_test.go covers pure backend plugin wiring helpers without starting the
// host HTTP server or scheduler.

package backend

import (
	"context"
	"strings"
	"testing"
	"time"

	"lina-core/pkg/plugin/capability/plugincap"
	feedingsvc "lina-plugin-sicau-niu/backend/internal/service/feeding"
)

type fakeIronLocationRefresher struct {
	called bool
}

func (f *fakeIronLocationRefresher) Refresh(ctx context.Context) (*feedingsvc.IronLocationRefreshResult, error) {
	f.called = true
	return &feedingsvc.IronLocationRefreshResult{}, nil
}

func TestBuildIronLocationRefreshDisabledWithoutCredentials(t *testing.T) {
	configSvc := newPluginTestConfigService(t, `
wechat:
  mock: true
`)

	enabled, refresher, _, err := buildIronLocationRefresh(context.Background(), configSvc)
	if err != nil {
		t.Fatalf("buildIronLocationRefresh returned error: %v", err)
	}
	if enabled || refresher != nil {
		t.Fatalf("expected refresh disabled without credentials")
	}
}

func TestBuildIronLocationRefreshDefaultsToOneMinute(t *testing.T) {
	configSvc := newPluginTestConfigService(t, `
niu:
  baseUrl: "https://example.test/iot"
  key: "key-1"
  secret: "secret-1"
`)

	enabled, refresher, interval, err := buildIronLocationRefresh(context.Background(), configSvc)
	if err != nil {
		t.Fatalf("buildIronLocationRefresh returned error: %v", err)
	}
	if !enabled || refresher == nil {
		t.Fatalf("expected IOT refresh to be enabled")
	}
	if interval != time.Minute {
		t.Fatalf("expected default interval 1min, got %s", interval)
	}
}

func TestBuildIronLocationRefreshClampsIntervalToOneMinute(t *testing.T) {
	configSvc := newPluginTestConfigService(t, `
niu:
  baseUrl: "https://example.test/iot"
  key: "key-1"
  secret: "secret-1"
  refreshInterval: 1s
`)

	enabled, refresher, interval, err := buildIronLocationRefresh(context.Background(), configSvc)
	if err != nil {
		t.Fatalf("buildIronLocationRefresh returned error: %v", err)
	}
	if !enabled || refresher == nil {
		t.Fatalf("expected IOT refresh to be enabled")
	}
	if interval != time.Minute {
		t.Fatalf("expected interval clamped to 1min, got %s", interval)
	}
}

func TestBuildIronLocationRefreshReadsBaseUrl(t *testing.T) {
	configSvc := newPluginTestConfigService(t, `
niu:
  baseUrl: "not-a-url"
  key: "key-1"
  secret: "secret-1"
`)

	_, _, _, err := buildIronLocationRefresh(context.Background(), configSvc)
	if err == nil {
		t.Fatalf("expected invalid baseUrl to fail")
	}
	if !strings.Contains(err.Error(), "niu.baseUrl") {
		t.Fatalf("expected baseUrl validation error, got %v", err)
	}
}

func TestRefreshIronLocationsSkipsNonPrimaryNode(t *testing.T) {
	refresher := &fakeIronLocationRefresher{}
	if err := refreshIronLocations(context.Background(), false, refresher); err != nil {
		t.Fatalf("refreshIronLocations returned error: %v", err)
	}
	if refresher.called {
		t.Fatalf("expected non-primary node to skip refresh")
	}
}

func TestRefreshIronLocationsUsesInjectedRefresherOnPrimaryNode(t *testing.T) {
	refresher := &fakeIronLocationRefresher{}
	if err := refreshIronLocations(context.Background(), true, refresher); err != nil {
		t.Fatalf("refreshIronLocations returned error: %v", err)
	}
	if !refresher.called {
		t.Fatalf("expected primary node to call refresher")
	}
}

func newPluginTestConfigService(t *testing.T, content string) plugincap.ConfigService {
	t.Helper()
	return plugincap.NewConfigFactory(t.TempDir(), t.TempDir()).
		WithArtifactConfig(pluginID, []byte(content)).
		ForPlugin(pluginID)
}
