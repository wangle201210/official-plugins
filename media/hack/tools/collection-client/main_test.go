// This file verifies the media collection TCP client flag normalization helpers.

package main

import (
	"context"
	"flag"
	"os"
	"testing"
	"time"

	"github.com/dellinger2023/net-flux/gen"
	"github.com/dellinger2023/net-flux/pkg/network"
)

// TestParseFlagsDefaultsToPersistentDiscovery verifies CLI defaults match server-side discovery.
func TestParseFlagsDefaultsToPersistentDiscovery(t *testing.T) {
	oldCommandLine := flag.CommandLine
	oldArgs := os.Args
	t.Cleanup(func() {
		flag.CommandLine = oldCommandLine
		os.Args = oldArgs
	})
	flag.CommandLine = flag.NewFlagSet("collection-client-test", flag.ContinueOnError)
	os.Args = []string{"collection-client"}

	cfg := parseFlags()
	if cfg.ephemeral {
		t.Fatal("expected discovery register packets to default to persistent instances")
	}
}

// TestNormalizeConfigFillsReportDefaults verifies report actions derive stable business keys.
func TestNormalizeConfigFillsReportDefaults(t *testing.T) {
	cfg, err := normalizeConfig(clientConfig{
		addr:        "127.0.0.1:1911",
		action:      actionReport,
		serviceName: "media-test",
		node:        2,
		privateIP:   "127.0.0.1",
		tenantID:    "tenant-a",
		timeout:     time.Second,
		protocol:    "http-flv",
		status:      "running",
	})
	if err != nil {
		t.Fatalf("normalize report config: %v", err)
	}
	if cfg.instanceID != "media-test" || cfg.nodeID != "node-2" {
		t.Fatalf("unexpected derived instance/node ids: %#v", cfg)
	}
	if cfg.streamID != "stream-media-test" || cfg.sessionID != "session-media-test" ||
		cfg.deviceID != "device-media-test" {
		t.Fatalf("unexpected derived report ids: %#v", cfg)
	}
	if cfg.streamPath != "rtmp://127.0.0.1/live/stream-media-test" {
		t.Fatalf("unexpected stream path: %q", cfg.streamPath)
	}
}

// TestNormalizeConfigRejectsInvalidAction verifies unsupported actions fail before connecting.
func TestNormalizeConfigRejectsInvalidAction(t *testing.T) {
	_, err := normalizeConfig(clientConfig{
		addr:        "127.0.0.1:1911",
		action:      "unknown",
		serviceName: "media-test",
		node:        1,
		timeout:     time.Second,
	})
	if err == nil {
		t.Fatal("expected invalid action error")
	}
}

// TestNormalizeConfigAllowsPingWithoutDiscoveryEndpoint verifies ping only needs TCP address and timeout.
func TestNormalizeConfigAllowsPingWithoutDiscoveryEndpoint(t *testing.T) {
	cfg, err := normalizeConfig(clientConfig{
		addr:        "127.0.0.1:1911",
		action:      actionPing,
		serviceName: "media-test",
		node:        1,
		timeout:     time.Second,
	})
	if err != nil {
		t.Fatalf("normalize ping config: %v", err)
	}
	if cfg.action != actionPing {
		t.Fatalf("unexpected action: %q", cfg.action)
	}
}

// TestClientHandlerSystemPackets verifies only Pong callbacks are accepted.
func TestClientHandlerSystemPackets(t *testing.T) {
	handler := newClientHandler(false)
	want := int64(123)
	if err := handler.OnCmdSystem(nil, &gen.Pong{Timestamp: want}); err != nil {
		t.Fatalf("handle pong: %v", err)
	}
	pong, err := handler.waitPong(context.Background())
	if err != nil {
		t.Fatalf("wait pong: %v", err)
	}
	if pong.GetTimestamp() != want {
		t.Fatalf("pong timestamp = %d, want %d", pong.GetTimestamp(), want)
	}
	if err := handler.OnCmdSystem(nil, &gen.LookupAck{}); err == nil {
		t.Fatal("expected unexpected system packet error")
	}
}

var _ network.EventHandler = (*clientHandler)(nil)

// TestParseStreamProtocolAcceptsCommonTokens verifies CLI protocol token mapping.
func TestParseStreamProtocolAcceptsCommonTokens(t *testing.T) {
	tests := map[string]gen.StreamProtocol{
		"http-flv": gen.StreamProtocol_SP_HTTP_FLV,
		"ws_flv":   gen.StreamProtocol_SP_WS_FLV,
		"gb28181":  gen.StreamProtocol_SP_GB28181,
	}
	for value, want := range tests {
		got, err := parseStreamProtocol(value)
		if err != nil {
			t.Fatalf("parse protocol %q: %v", value, err)
		}
		if got != want {
			t.Fatalf("protocol %q = %s, want %s", value, got.String(), want.String())
		}
	}
}

// TestParseStreamStatusAcceptsCommonTokens verifies CLI status token mapping.
func TestParseStreamStatusAcceptsCommonTokens(t *testing.T) {
	tests := map[string]gen.StreamStatus{
		"running":   gen.StreamStatus_SS_RUNNING,
		"cancelled": gen.StreamStatus_SS_CANCELLED,
		"canceled":  gen.StreamStatus_SS_CANCELLED,
		"failed":    gen.StreamStatus_SS_FAILED,
	}
	for value, want := range tests {
		got, err := parseStreamStatus(value)
		if err != nil {
			t.Fatalf("parse status %q: %v", value, err)
		}
		if got != want {
			t.Fatalf("status %q = %s, want %s", value, got.String(), want.String())
		}
	}
}
