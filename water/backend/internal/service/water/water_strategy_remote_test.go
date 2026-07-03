// This file tests remote media strategy resolution over HTTP.

package water

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"lina-core/pkg/bizerr"
)

// TestRemoteStrategyResolverCallsMediaOpenAPI verifies remote strategy lookup request shape.
func TestRemoteStrategyResolverCallsMediaOpenAPI(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != strategyResolverPath {
			t.Fatalf("expected path %s, got %s", strategyResolverPath, r.URL.Path)
		}
		if r.Header.Get(strategyResolverAPIKeyHeader) != "secret" {
			t.Fatal("expected inner api key header")
		}
		if r.URL.Query().Get("tenantId") != "tenant-a" {
			t.Fatalf("expected tenant query, got %q", r.URL.Query().Get("tenantId"))
		}
		if r.URL.Query().Get("deviceId") != "device-a" {
			t.Fatalf("expected device query, got %q", r.URL.Query().Get("deviceId"))
		}
		w.Header().Set("Content-Type", "application/json")
		if _, err := w.Write([]byte(`{
			"code": 0,
			"message": "OK",
			"data": {
				"matched": true,
				"source": "device",
				"sourceLabel": "设备策略",
				"strategyId": 17,
				"strategyName": "remote strategy",
				"strategy": "snapshot_watermark:\n  order: Remote"
			}
		}`)); err != nil {
			t.Fatalf("write response: %v", err)
		}
	}))
	defer server.Close()

	resolver, err := NewRemoteStrategyResolver(RemoteStrategyResolverConfig{
		BaseURL: server.URL,
		APIKey:  "secret",
		Timeout: time.Second,
	})
	if err != nil {
		t.Fatalf("new remote resolver failed: %v", err)
	}
	out, err := resolver.ResolveStrategy(context.Background(), ResolveStrategyInput{
		TenantId: " tenant-a ",
		DeviceId: " device-a ",
	})
	if err != nil {
		t.Fatalf("resolve strategy failed: %v", err)
	}
	if out == nil || !out.Matched {
		t.Fatalf("expected matched strategy, got %#v", out)
	}
	if out.StrategyId != 17 || out.Source != string(StrategySourceDevice) {
		t.Fatalf("expected remote strategy metadata, got %#v", out)
	}
	if out.Strategy != "snapshot_watermark:\n  order: Remote" {
		t.Fatalf("expected strategy body to roundtrip, got %q", out.Strategy)
	}
}

// TestRemoteStrategyResolverDecodesDirectPayload verifies compatibility with unwrapped responses.
func TestRemoteStrategyResolverDecodesDirectPayload(t *testing.T) {
	out, err := decodeRemoteStrategyResponse([]byte(`{
		"matched": false,
		"source": "none",
		"sourceLabel": "未匹配"
	}`))
	if err != nil {
		t.Fatalf("decode direct payload failed: %v", err)
	}
	if out == nil || out.Matched || out.Source != string(StrategySourceNone) {
		t.Fatalf("expected direct none payload, got %#v", out)
	}
}

// TestDisabledRemoteStrategyResolverReturnsUnavailable verifies missing config fails at call time.
func TestDisabledRemoteStrategyResolverReturnsUnavailable(t *testing.T) {
	resolver, err := NewRemoteStrategyResolver(RemoteStrategyResolverConfig{})
	if err != nil {
		t.Fatalf("new disabled resolver failed: %v", err)
	}
	_, err = resolver.ResolveStrategy(context.Background(), ResolveStrategyInput{})
	if !bizerr.Is(err, CodeWaterMediaResolverUnavailable) {
		t.Fatalf("expected unavailable error, got %v", err)
	}
}
