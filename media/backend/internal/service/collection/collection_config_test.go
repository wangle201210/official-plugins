// This file verifies media collection server configuration loading.

package collection

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/gogf/gf/v2/container/gvar"

	"gopkg.in/yaml.v3"
	"lina-core/pkg/plugin/capability/plugincap"
)

// TestLoadConfigUsesDefaultsWhenUnset verifies collection defaults when config is absent.
func TestLoadConfigUsesDefaultsWhenUnset(t *testing.T) {
	cfg, err := LoadConfig(context.Background(), newTestConfigService(t, "media:\n  placeholder: true\n"))
	if err != nil {
		t.Fatalf("load collection config: %v", err)
	}
	if cfg.Enabled {
		t.Fatal("expected collection server disabled by default")
	}
	if cfg.Addr != defaultAddr {
		t.Fatalf("expected default addr %s, got %s", defaultAddr, cfg.Addr)
	}
	if cfg.Discovery.Enabled {
		t.Fatal("expected discovery disabled by default")
	}
	if cfg.Discovery.Host != defaultDiscoveryHost {
		t.Fatalf("expected default discovery host %s, got %s", defaultDiscoveryHost, cfg.Discovery.Host)
	}
	if cfg.Discovery.Port != defaultDiscoveryPort {
		t.Fatalf("expected default discovery port %d, got %d", defaultDiscoveryPort, cfg.Discovery.Port)
	}
}

// TestLoadConfigUsesConfiguredValues verifies configured collection values override defaults.
func TestLoadConfigUsesConfiguredValues(t *testing.T) {
	cfg, err := LoadConfig(context.Background(), newTestConfigService(t, `
collectionServer:
  enabled: true
  addr: "127.0.0.1:1912"
  discovery:
    enabled: true
    host: "10.0.0.8"
    port: 8849
    namespace: "media"
    logDir: "./media-logs"
    cacheDir: "./media-cache"
    notLoadCacheAtStart: false
    timeout: 3000
    username: "media-user"
    password: "media-pass"
    node: 7
`))
	if err != nil {
		t.Fatalf("load collection config: %v", err)
	}
	if !cfg.Enabled {
		t.Fatal("expected collection server enabled")
	}
	if cfg.Addr != "127.0.0.1:1912" {
		t.Fatalf("expected configured addr, got %s", cfg.Addr)
	}
	if !cfg.Discovery.Enabled {
		t.Fatal("expected discovery enabled")
	}
	if cfg.Discovery.Host != "10.0.0.8" {
		t.Fatalf("expected configured discovery host, got %s", cfg.Discovery.Host)
	}
	if cfg.Discovery.Port != 8849 {
		t.Fatalf("expected configured discovery port, got %d", cfg.Discovery.Port)
	}
	if cfg.Discovery.Namespace != "media" {
		t.Fatalf("expected configured discovery namespace, got %s", cfg.Discovery.Namespace)
	}
	if cfg.Discovery.LogDir != "./media-logs" {
		t.Fatalf("expected configured discovery log dir, got %s", cfg.Discovery.LogDir)
	}
	if cfg.Discovery.CacheDir != "./media-cache" {
		t.Fatalf("expected configured discovery cache dir, got %s", cfg.Discovery.CacheDir)
	}
	if cfg.Discovery.NotLoadCacheAtStart {
		t.Fatal("expected configured discovery notLoadCacheAtStart false")
	}
	if cfg.Discovery.Timeout != 3000 {
		t.Fatalf("expected configured discovery timeout, got %d", cfg.Discovery.Timeout)
	}
	if cfg.Discovery.Username != "media-user" {
		t.Fatalf("expected configured discovery username, got %s", cfg.Discovery.Username)
	}
	if cfg.Discovery.Password != "media-pass" {
		t.Fatalf("expected configured discovery password, got %s", cfg.Discovery.Password)
	}
	if cfg.Discovery.Node != 7 {
		t.Fatalf("expected configured discovery node, got %d", cfg.Discovery.Node)
	}
}

// TestLoadConfigRejectsBlankAddress verifies empty listen addresses are rejected.
func TestLoadConfigRejectsBlankAddress(t *testing.T) {
	_, err := LoadConfig(context.Background(), newTestConfigService(t, `
collectionServer:
  enabled: true
  addr: "   "
`))
	if err == nil {
		t.Fatal("expected blank address error")
	}
	if !strings.Contains(err.Error(), configKeyCollectionServerAddr) {
		t.Fatalf("expected error to mention %s, got %v", configKeyCollectionServerAddr, err)
	}
}

// TestLoadConfigRejectsInvalidDiscovery verifies enabled discovery validates Nacos settings.
func TestLoadConfigRejectsInvalidDiscovery(t *testing.T) {
	_, err := LoadConfig(context.Background(), newTestConfigService(t, `
collectionServer:
  discovery:
    enabled: true
    port: 0
`))
	if err == nil {
		t.Fatal("expected invalid discovery port error")
	}
	if !strings.Contains(err.Error(), configKeyCollectionServerDiscoveryPort) {
		t.Fatalf("expected error to mention %s, got %v", configKeyCollectionServerDiscoveryPort, err)
	}
}

// newTestConfigService builds a scoped plugin config reader from artifact content.
func newTestConfigService(t *testing.T, content string) plugincap.ConfigService {
	t.Helper()

	var values map[string]any
	if err := yaml.Unmarshal([]byte(content), &values); err != nil {
		t.Fatalf("parse collection config test data: %v", err)
	}
	return collectionConfigTestService{values: values}
}

type collectionConfigTestService struct {
	values map[string]any
}

func (s collectionConfigTestService) Get(_ context.Context, key string, defaultValue any) (*gvar.Var, error) {
	value, ok := lookupCollectionConfigTestValue(s.values, key)
	if !ok {
		if defaultValue == nil {
			return nil, nil
		}
		return gvar.New(defaultValue), nil
	}
	return gvar.New(value), nil
}

func (s collectionConfigTestService) Exists(_ context.Context, key string) (bool, error) {
	_, ok := lookupCollectionConfigTestValue(s.values, key)
	return ok, nil
}

func (s collectionConfigTestService) Scan(_ context.Context, key string, target any) error {
	value, ok := lookupCollectionConfigTestValue(s.values, key)
	if !ok {
		return nil
	}
	payload, err := yaml.Marshal(value)
	if err != nil {
		return err
	}
	return yaml.Unmarshal(payload, target)
}

func (s collectionConfigTestService) String(ctx context.Context, key string, defaultValue string) (string, error) {
	value, err := s.Get(ctx, key, defaultValue)
	if err != nil {
		return "", err
	}
	if value == nil || value.IsNil() {
		return defaultValue, nil
	}
	raw := value.String()
	if strings.TrimSpace(raw) == "" {
		return defaultValue, nil
	}
	return raw, nil
}

func (s collectionConfigTestService) Bool(ctx context.Context, key string, defaultValue bool) (bool, error) {
	value, err := s.Get(ctx, key, defaultValue)
	if err != nil {
		return false, err
	}
	if value == nil || value.IsNil() {
		return defaultValue, nil
	}
	return value.Bool(), nil
}

func (s collectionConfigTestService) Int(ctx context.Context, key string, defaultValue int) (int, error) {
	value, err := s.Get(ctx, key, defaultValue)
	if err != nil {
		return 0, err
	}
	if value == nil || value.IsNil() {
		return defaultValue, nil
	}
	return value.Int(), nil
}

func (s collectionConfigTestService) Duration(ctx context.Context, key string, defaultValue time.Duration) (time.Duration, error) {
	value, err := s.Get(ctx, key, defaultValue)
	if err != nil {
		return 0, err
	}
	if value == nil || value.IsNil() {
		return defaultValue, nil
	}
	raw := strings.TrimSpace(value.String())
	if raw == "" {
		return defaultValue, nil
	}
	duration, err := time.ParseDuration(raw)
	if err != nil {
		return 0, err
	}
	return duration, nil
}

func lookupCollectionConfigTestValue(values map[string]any, key string) (any, bool) {
	if values == nil {
		return nil, false
	}
	current := any(values)
	for _, part := range strings.Split(key, ".") {
		mapped, ok := current.(map[string]any)
		if !ok {
			return nil, false
		}
		current, ok = mapped[part]
		if !ok {
			return nil, false
		}
	}
	return current, true
}
