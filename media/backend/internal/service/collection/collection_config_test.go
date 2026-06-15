// This file verifies media collection server configuration loading.

package collection

import (
	"context"
	"strings"
	"testing"

	"lina-core/pkg/plugin/capability/plugincap"
	configsvc "lina-core/pkg/plugin/capability/plugincap"
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
    preloadCache: false
    timeout: 3000
    groupName: "MEDIA_GROUP"
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
	if cfg.Discovery.PreloadCache {
		t.Fatal("expected configured discovery preload cache false")
	}
	if cfg.Discovery.Timeout != 3000 {
		t.Fatalf("expected configured discovery timeout, got %d", cfg.Discovery.Timeout)
	}
	if cfg.Discovery.GroupName != "MEDIA_GROUP" {
		t.Fatalf("expected configured discovery group name, got %s", cfg.Discovery.GroupName)
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

	return configsvc.NewConfigFactory(t.TempDir(), t.TempDir()).
		WithArtifactConfig("media", []byte(content)).
		ForPlugin("media")
}
