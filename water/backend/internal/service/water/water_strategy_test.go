// This file tests watermark strategy parsing.

package water

import (
	"context"
	"testing"
)

// TestParseWatermarkStrategySnapshotNode verifies Lina snapshot watermark YAML parsing.
func TestParseWatermarkStrategySnapshotNode(t *testing.T) {
	cfg, err := parseWatermarkStrategy(`record:
  enabled: true
snapshot_watermark:
  order: 园区安防
  font_size: 48
  color: "#00ff88"
  opacity: 0.5
  width: 1280
  height: 720
  rotate: -30`)
	if err != nil {
		t.Fatalf("parse watermark strategy failed: %v", err)
	}
	if cfg == nil {
		t.Fatal("expected watermark config")
	}
	if cfg.Order != "园区安防" {
		t.Fatalf("expected order to roundtrip, got %q", cfg.Order)
	}
	if cfg.FontSize != 48 {
		t.Fatalf("expected font size 48, got %d", cfg.FontSize)
	}
	if cfg.Width != 1280 || cfg.Height != 720 {
		t.Fatalf("expected dimensions 1280x720, got %dx%d", cfg.Width, cfg.Height)
	}
	if cfg.Rotate != -30 {
		t.Fatalf("expected rotate -30, got %d", cfg.Rotate)
	}
}

// TestParseWatermarkStrategyConvertsLibraryConfig verifies service config maps to the renderer config.
func TestParseWatermarkStrategyConvertsLibraryConfig(t *testing.T) {
	cfg, err := parseWatermarkStrategy(`snapshot_watermark:
  order: 热点水印
  font_size: 64
  opacity: 0.15
  width: 640
  height: 360
  base64: "data:image/png;base64,aGVsbG8="
  rotate: 15`)
	if err != nil {
		t.Fatalf("parse watermark strategy failed: %v", err)
	}
	if cfg == nil {
		t.Fatal("expected snapshot watermark config")
	}
	if cfg.Opacity != 0.15 {
		t.Fatalf("expected opacity 0.15, got %f", cfg.Opacity)
	}
	rendererCfg := cfg.ToWatermarkConfig()
	if rendererCfg.Order != "热点水印" || rendererCfg.FontSize != 64 {
		t.Fatalf("expected renderer text config to roundtrip, got %+v", rendererCfg)
	}
	if rendererCfg.Width != 640 || rendererCfg.Height != 360 || rendererCfg.Rotate != 15 {
		t.Fatalf("expected renderer dimensions and rotation to roundtrip, got %+v", rendererCfg)
	}
	if rendererCfg.Base64 != "aGVsbG8=" {
		t.Fatalf("expected renderer base64 without data URL prefix, got %q", rendererCfg.Base64)
	}
}

// TestParseWatermarkStrategyDefaultOpacity verifies omitted opacity uses the snapshot watermark default.
func TestParseWatermarkStrategyDefaultOpacity(t *testing.T) {
	cfg, err := parseWatermarkStrategy(`snapshot_watermark:
  order: 默认透明度`)
	if err != nil {
		t.Fatalf("parse watermark strategy failed: %v", err)
	}
	if cfg == nil {
		t.Fatal("expected snapshot watermark config")
	}
	if cfg.Opacity != 0.15 {
		t.Fatalf("expected default opacity 0.15, got %f", cfg.Opacity)
	}
}

// TestParseWatermarkStrategyMissing verifies strategies without snapshot watermark are skipped.
func TestParseWatermarkStrategyMissing(t *testing.T) {
	cfg, err := parseWatermarkStrategy(`record:
  enabled: true`)
	if err != nil {
		t.Fatalf("parse watermark strategy failed: %v", err)
	}
	if cfg != nil {
		t.Fatalf("expected nil watermark config, got %+v", cfg)
	}
}

// TestParseWatermarkStrategyIgnoresGenericWatermark verifies only screenshot-specific nodes are accepted.
func TestParseWatermarkStrategyIgnoresGenericWatermark(t *testing.T) {
	cfg, err := parseWatermarkStrategy(`watermark:
  enabled: true
  order: 通用水印`)
	if err != nil {
		t.Fatalf("parse watermark strategy failed: %v", err)
	}
	if cfg != nil {
		t.Fatalf("expected generic watermark node to be ignored, got %+v", cfg)
	}
}

// TestResolveStrategyCachesTenantDeviceResult verifies repeated lookups reuse host cache.
func TestResolveStrategyCachesTenantDeviceResult(t *testing.T) {
	ctx := context.Background()
	cacheSvc := newTaskStoreCache()
	resolver := &countingStrategyResolver{
		next: &ResolveStrategyOutput{
			Matched:      true,
			Source:       string(StrategySourceTenantDevice),
			SourceLabel:  strategySourceLabel(StrategySourceTenantDevice),
			StrategyId:   17,
			StrategyName: "租户设备策略",
			Strategy:     "snapshot_watermark:\n  order: cached\n",
		},
	}
	service := &serviceImpl{strategyCache: cacheSvc, strategyResolver: resolver}

	first, err := service.resolveStrategy(ctx, " tenant-a ", " device-a ")
	if err != nil {
		t.Fatalf("resolve first strategy: %v", err)
	}
	resolver.next = &ResolveStrategyOutput{
		Matched:      true,
		Source:       string(StrategySourceDevice),
		SourceLabel:  strategySourceLabel(StrategySourceDevice),
		StrategyId:   23,
		StrategyName: "设备策略",
		Strategy:     "snapshot_watermark:\n  order: changed\n",
	}
	second, err := service.resolveStrategy(ctx, "tenant-a", "device-a")
	if err != nil {
		t.Fatalf("resolve cached strategy: %v", err)
	}

	if resolver.calls != 1 {
		t.Fatalf("expected one resolver call after cache hit, got %d", resolver.calls)
	}
	if first.StrategyId != 17 || second.StrategyId != 17 {
		t.Fatalf("expected cached strategy id 17, first=%+v second=%+v", first, second)
	}
	if cacheSvc.lastNamespace != strategyResolveCacheNamespace {
		t.Fatalf("expected strategy cache namespace, got %q", cacheSvc.lastNamespace)
	}
	if cacheSvc.lastKey != "water:strategy:tenant:tenant-a:device:device-a" {
		t.Fatalf("expected readable tenant-device cache key, got %q", cacheSvc.lastKey)
	}
	if cacheSvc.lastTTL != strategyResolveCacheTTL {
		t.Fatalf("expected strategy cache TTL %s, got %s", strategyResolveCacheTTL, cacheSvc.lastTTL)
	}
}

// TestStrategyResolveCacheKeyAllowsEmptyDevice verifies tenant-only lookups use a readable blank segment.
func TestStrategyResolveCacheKeyAllowsEmptyDevice(t *testing.T) {
	key := strategyResolveCacheKey(ResolveStrategyInput{TenantId: " tenant-a ", DeviceId: " "})
	if key != "water:strategy:tenant:tenant-a:device:_" {
		t.Fatalf("expected readable tenant-only key, got %q", key)
	}
}

type countingStrategyResolver struct {
	calls int
	next  *ResolveStrategyOutput
}

func (r *countingStrategyResolver) ResolveStrategy(_ context.Context, _ ResolveStrategyInput) (*ResolveStrategyOutput, error) {
	r.calls++
	return r.next, nil
}
