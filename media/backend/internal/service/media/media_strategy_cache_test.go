// This file tests short-lived media strategy resolution caching.

package media

import (
	"context"
	"strings"
	"testing"

	"lina-core/pkg/plugin/capability/bizctxcap"
	"lina-plugin-media/backend/internal/dao"
	"lina-plugin-media/backend/internal/model/do"
)

// TestResolveStrategyCachesTenantDeviceResult verifies repeated effective strategy lookups reuse cache.
func TestResolveStrategyCachesTenantDeviceResult(t *testing.T) {
	ctx := context.Background()
	setupMediaStrategySQLite(t, ctx)
	cacheSvc := newMemoryRouteMemoryCache()
	svc, err := newWithRouteMemoryCache(bizctxcap.New(nil), cacheSvc)
	if err != nil {
		t.Fatalf("create media service: %v", err)
	}

	firstStrategyID := insertTestStrategy(t, ctx, "租户设备策略 A", int(SwitchOff), int(SwitchOn))
	secondStrategyID := insertTestStrategy(t, ctx, "租户设备策略 B", int(SwitchOff), int(SwitchOn))
	if _, err = dao.MediaStrategyDeviceTenant.Ctx(ctx).Data(do.MediaStrategyDeviceTenant{
		TenantId:   "tenant-a",
		DeviceId:   "device-a",
		StrategyId: firstStrategyID,
	}).Insert(); err != nil {
		t.Fatalf("insert tenant-device binding: %v", err)
	}

	first, err := svc.ResolveStrategy(ctx, ResolveStrategyInput{TenantId: "tenant-a", DeviceId: "device-a"})
	if err != nil {
		t.Fatalf("resolve first strategy: %v", err)
	}
	if _, err = dao.MediaStrategyDeviceTenant.Ctx(ctx).
		Where(do.MediaStrategyDeviceTenant{TenantId: "tenant-a", DeviceId: "device-a"}).
		Data(do.MediaStrategyDeviceTenant{StrategyId: secondStrategyID}).
		Update(); err != nil {
		t.Fatalf("update tenant-device binding outside service: %v", err)
	}
	second, err := svc.ResolveStrategy(ctx, ResolveStrategyInput{TenantId: " tenant-a ", DeviceId: " device-a "})
	if err != nil {
		t.Fatalf("resolve cached strategy: %v", err)
	}

	if first.StrategyId != firstStrategyID || second.StrategyId != firstStrategyID {
		t.Fatalf("expected cached strategy id %d, first=%+v second=%+v", firstStrategyID, first, second)
	}
	expectedKey := "resolve:tenant:tenant-a:device:device-a:version:0"
	if cacheSvc.lastNamespace != strategyResolveCacheNamespace || cacheSvc.lastKey != expectedKey {
		t.Fatalf("expected readable strategy cache key %s/%s, got %s/%s",
			strategyResolveCacheNamespace, expectedKey, cacheSvc.lastNamespace, cacheSvc.lastKey)
	}
	if strings.Contains(cacheSvc.lastKey, "sha256") {
		t.Fatalf("expected readable cache key without hash marker, got %q", cacheSvc.lastKey)
	}
	if cacheSvc.lastTTL != strategyResolveCacheTTL {
		t.Fatalf("expected strategy cache TTL %s, got %s", strategyResolveCacheTTL, cacheSvc.lastTTL)
	}
}

// TestResolveStrategyInvalidatesCacheAfterBindingMutation verifies service mutations bypass older cache entries.
func TestResolveStrategyInvalidatesCacheAfterBindingMutation(t *testing.T) {
	ctx := context.Background()
	setupMediaStrategySQLite(t, ctx)
	cacheSvc := newMemoryRouteMemoryCache()
	svc, err := newWithRouteMemoryCache(bizctxcap.New(nil), cacheSvc)
	if err != nil {
		t.Fatalf("create media service: %v", err)
	}

	firstStrategyID := insertTestStrategy(t, ctx, "租户设备策略 A", int(SwitchOff), int(SwitchOn))
	secondStrategyID := insertTestStrategy(t, ctx, "租户设备策略 B", int(SwitchOff), int(SwitchOn))
	if _, err = svc.SaveTenantDeviceBinding(ctx, TenantDeviceBindingMutationInput{
		TenantId:   "tenant-a",
		DeviceId:   "device-a",
		StrategyId: firstStrategyID,
	}); err != nil {
		t.Fatalf("save first tenant-device binding: %v", err)
	}
	first, err := svc.ResolveStrategy(ctx, ResolveStrategyInput{TenantId: "tenant-a", DeviceId: "device-a"})
	if err != nil {
		t.Fatalf("resolve first strategy: %v", err)
	}
	if first.StrategyId != firstStrategyID {
		t.Fatalf("expected first strategy id %d, got %+v", firstStrategyID, first)
	}

	if _, err = svc.SaveTenantDeviceBinding(ctx, TenantDeviceBindingMutationInput{
		TenantId:   "tenant-a",
		DeviceId:   "device-a",
		StrategyId: secondStrategyID,
	}); err != nil {
		t.Fatalf("save second tenant-device binding: %v", err)
	}
	second, err := svc.ResolveStrategy(ctx, ResolveStrategyInput{TenantId: "tenant-a", DeviceId: "device-a"})
	if err != nil {
		t.Fatalf("resolve invalidated strategy: %v", err)
	}
	if second.StrategyId != secondStrategyID {
		t.Fatalf("expected second strategy id %d after invalidation, got %+v", secondStrategyID, second)
	}
}

// TestStrategyResolveCacheKeyAllowsEmptyDevice verifies tenant-only lookups use a readable blank segment.
func TestStrategyResolveCacheKeyAllowsEmptyDevice(t *testing.T) {
	key := strategyResolveCacheKey(7, ResolveStrategyInput{TenantId: " tenant-a ", DeviceId: " "})
	if key != "resolve:tenant:tenant-a:device:_:version:7" {
		t.Fatalf("expected readable tenant-only key, got %q", key)
	}
}
