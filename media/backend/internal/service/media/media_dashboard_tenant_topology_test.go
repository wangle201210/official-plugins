// This file verifies tenant media-node topology aggregation and validation.

package media

import (
	"context"
	"strings"
	"testing"

	"lina-plugin-media/backend/internal/model/do"
)

// TestDashboardTenantTopologyAggregatesReportedData verifies tenant and node statistics.
func TestDashboardTenantTopologyAggregatesReportedData(t *testing.T) {
	ctx := context.Background()
	setupMediaStrategySQLite(t, ctx)
	setupMediaDashboardReportTables(t, ctx)
	insertDashboardReportFixtures(t, ctx)

	insertDashboardReports(t, ctx, []any{
		do.MediaReportStream{
			StreamId:     "stream-d",
			TenantId:     "tenant-a",
			NodeId:       "node-b",
			NodeName:     "Stale Node B",
			ProtocolType: "RTMP",
			Status:       "playing",
			ReportTime:   int64(1780000000001),
		},
		do.MediaReportStream{
			StreamId:     "stream-e",
			TenantId:     "tenant-a",
			NodeId:       "node-b",
			NodeName:     "Stale Node B",
			ProtocolType: "RTSP",
			Status:       "playing",
			ReportTime:   int64(1780000000002),
		},
	})
	session := dashboardSessionDO(
		"session-e",
		"stream-d",
		"Camera D",
		"tenant-a",
		"client-e",
		"viewer-e",
		"RTMP",
		"inst-c",
		30,
		1780000000001,
		nil,
	)
	session.NodeId = "node-b"
	session.NodeName = "Stale Node B"
	insertDashboardReports(t, ctx, []any{session})

	topology, err := newTestMediaService(t).GetDashboardTenantTopology(ctx, DashboardTenantTopologyInput{
		TenantId: " tenant-a ",
	})
	if err != nil {
		t.Fatalf("get dashboard tenant topology: %v", err)
	}
	if topology.TenantId != "tenant-a" || topology.ConcurrentSessions != 4 ||
		topology.LiveStreamCount != 3 || topology.ProtocolCount != 3 ||
		topology.NodeCount != 2 || topology.NodeDetailLimited || topology.GeneratedAt <= 0 {
		t.Fatalf("unexpected tenant totals %#v", topology)
	}
	assertDashboardTenantProtocols(t, topology.Protocols, map[string]int{
		"HLS":  1,
		"RTMP": 1,
		"RTSP": 1,
	})
	if len(topology.Nodes) != 2 {
		t.Fatalf("expected two nodes, got %#v", topology.Nodes)
	}
	assertDashboardTenantNode(t, topology.Nodes[0], "node-a", "Node A", 3, 1, map[string]int{"HLS": 1})
	assertDashboardTenantNode(t, topology.Nodes[1], "node-b", "Node B", 1, 2, map[string]int{"RTMP": 1, "RTSP": 1})

	empty, err := newTestMediaService(t).GetDashboardTenantTopology(ctx, DashboardTenantTopologyInput{
		TenantId: "missing-tenant",
	})
	if err != nil {
		t.Fatalf("get empty tenant topology: %v", err)
	}
	if empty.ConcurrentSessions != 0 || empty.LiveStreamCount != 0 ||
		len(empty.Protocols) != 0 || len(empty.Nodes) != 0 {
		t.Fatalf("expected empty tenant topology, got %#v", empty)
	}
}

// TestDashboardTenantTopologyRejectsInvalidTenantID verifies service-level validation.
func TestDashboardTenantTopologyRejectsInvalidTenantID(t *testing.T) {
	ctx := context.Background()
	setupMediaStrategySQLite(t, ctx)
	setupMediaDashboardReportTables(t, ctx)
	service := newTestMediaService(t)

	if _, err := service.GetDashboardTenantTopology(ctx, DashboardTenantTopologyInput{}); err == nil {
		t.Fatal("expected missing tenant ID to be rejected")
	}
	if _, err := service.GetDashboardTenantTopology(ctx, DashboardTenantTopologyInput{
		TenantId: strings.Repeat("a", 65),
	}); err == nil {
		t.Fatal("expected overlong tenant ID to be rejected")
	}
}

// TestDashboardTenantTopologyEnforcesTietaTenant verifies fallback auth cannot cross tenant boundaries.
func TestDashboardTenantTopologyEnforcesTietaTenant(t *testing.T) {
	ctx := context.Background()
	setupMediaStrategySQLite(t, ctx)
	setupMediaDashboardReportTables(t, ctx)
	service := newTestMediaService(t)

	if _, err := service.GetDashboardTenantTopology(
		WithDashboardTietaTenantContext(ctx, "tenant-a"),
		DashboardTenantTopologyInput{TenantId: "tenant-b"},
	); err == nil {
		t.Fatal("expected mismatched Tieta tenant to be rejected")
	}
	if _, err := service.GetDashboardTenantTopology(
		WithDashboardTietaTenantContext(ctx, ""),
		DashboardTenantTopologyInput{TenantId: "tenant-a"},
	); err == nil {
		t.Fatal("expected missing Tieta tenant to be rejected")
	}
	if _, err := service.GetDashboardTenantTopology(
		WithDashboardTietaTenantContext(ctx, "tenant-a"),
		DashboardTenantTopologyInput{TenantId: "tenant-a"},
	); err != nil {
		t.Fatalf("expected matching Tieta tenant to be accepted: %v", err)
	}
}

func assertDashboardTenantNode(
	t *testing.T,
	node *DashboardTenantTopologyNode,
	nodeID string,
	nodeName string,
	concurrentSessions int,
	liveStreamCount int,
	protocols map[string]int,
) {
	t.Helper()
	if node == nil || node.NodeId != nodeID || node.NodeName != nodeName ||
		node.ConcurrentSessions != concurrentSessions || node.LiveStreamCount != liveStreamCount ||
		node.ProtocolCount != len(protocols) {
		t.Fatalf("unexpected tenant node %#v", node)
	}
	assertDashboardTenantProtocols(t, node.Protocols, protocols)
}

func assertDashboardTenantProtocols(
	t *testing.T,
	items []*DashboardTenantTopologyProtocol,
	expected map[string]int,
) {
	t.Helper()
	if len(items) != len(expected) {
		t.Fatalf("expected protocols %#v, got %#v", expected, items)
	}
	for _, item := range items {
		if item == nil {
			t.Fatalf("expected protocols %#v, got nil item in %#v", expected, items)
		}
		expectedCount, ok := expected[item.ProtocolType]
		if !ok || expectedCount != item.LiveStreamCount {
			t.Fatalf("expected protocols %#v, got %#v", expected, items)
		}
	}
}
