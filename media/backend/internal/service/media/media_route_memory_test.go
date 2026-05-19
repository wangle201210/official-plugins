// This file tests HotGo-compatible route memory.

package media

import (
	"context"
	"testing"
	"time"

	"lina-core/pkg/pluginservice/bizctx"
)

// TestRouteMemoryUsesDeviceChannelKeyAndTwelveHourTTL verifies route memory lifecycle behavior.
func TestRouteMemoryUsesDeviceChannelKeyAndTwelveHourTTL(t *testing.T) {
	ctx := context.Background()
	cacheSvc := newMemoryRouteMemoryCache()
	svc, err := newWithRouteMemoryCache(bizctx.New(nil), cacheSvc)
	if err != nil {
		t.Fatalf("create media service: %v", err)
	}

	if err = svc.SetRouteMemory(ctx, RouteMemoryInput{
		RouteMemoryKeyInput: RouteMemoryKeyInput{
			DeviceCode:  "34020000001320000001",
			ChannelCode: "34020000001320000002",
		},
		Data: "node-a",
	}); err != nil {
		t.Fatalf("set route memory: %v", err)
	}
	if cacheSvc.lastNamespace != "route-memory" {
		t.Fatalf("expected route-memory namespace, got %q", cacheSvc.lastNamespace)
	}
	if cacheSvc.lastKey != "route_data:34020000001320000001:34020000001320000002" {
		t.Fatalf("expected HotGo route key, got %q", cacheSvc.lastKey)
	}
	if cacheSvc.lastTTL != 12*time.Hour {
		t.Fatalf("expected 12h TTL, got %s", cacheSvc.lastTTL)
	}

	out, err := svc.GetRouteMemory(ctx, RouteMemoryKeyInput{
		DeviceCode:  "34020000001320000001",
		ChannelCode: "34020000001320000002",
	})
	if err != nil {
		t.Fatalf("get route memory: %v", err)
	}
	if out == nil || out.Data != "node-a" {
		t.Fatalf("expected stored route data, got %#v", out)
	}

	other, err := svc.GetRouteMemory(ctx, RouteMemoryKeyInput{
		DeviceCode:  "34020000001320000001",
		ChannelCode: "34020000001320000003",
	})
	if err != nil {
		t.Fatalf("get other route memory: %v", err)
	}
	if other == nil || other.Data != "" {
		t.Fatalf("expected other channel to miss, got %#v", other)
	}

	if err = svc.DeleteRouteMemory(ctx, RouteMemoryKeyInput{
		DeviceCode:  "34020000001320000001",
		ChannelCode: "34020000001320000002",
	}); err != nil {
		t.Fatalf("delete route memory: %v", err)
	}
	deleted, err := svc.GetRouteMemory(ctx, RouteMemoryKeyInput{
		DeviceCode:  "34020000001320000001",
		ChannelCode: "34020000001320000002",
	})
	if err != nil {
		t.Fatalf("get deleted route memory: %v", err)
	}
	if deleted == nil || deleted.Data != "" {
		t.Fatalf("expected deleted route memory to be empty, got %#v", deleted)
	}
}
