// This file verifies dashboard read models exposed by the media service.

package media

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/os/gtime"

	"lina-plugin-media/backend/internal/dao"
	"lina-plugin-media/backend/internal/model/do"
)

// TestDashboardQueriesReadReportProjections verifies dashboard queries use report tables and decode nested JSON.
func TestDashboardQueriesReadReportProjections(t *testing.T) {
	ctx := context.Background()
	setupMediaStrategySQLite(t, ctx)
	setupMediaDashboardReportTables(t, ctx)
	insertDashboardReportFixtures(t, ctx)

	svc := newTestMediaService(t)
	node, err := svc.GetDashboardNodeOverview(ctx, DashboardNodeOverviewInput{RootNodeId: "node-a"})
	if err != nil {
		t.Fatalf("get node overview: %v", err)
	}
	if node.Node == nil || node.Node.NodeId != "node-a" {
		t.Fatalf("expected node-a overview, got %#v", node)
	}
	if len(node.Node.ChildNodes) != 2 || node.Node.ChildNodes[0].NodeId != "node-b" || node.Node.ChildNodes[1].NodeId != "node-c" {
		t.Fatalf("expected node-b child, got %#v", node.Node.ChildNodes)
	}
	if len(node.Node.ChildNodes[0].ChildNodes) != 1 || node.Node.ChildNodes[0].ChildNodes[0].NodeId != "node-d" {
		t.Fatalf("expected node-d grandchild under node-b, got %#v", node.Node.ChildNodes[0].ChildNodes)
	}
	if node.Node.NodeLatencyMap["node-b"] != 17 || node.Node.NodeLatencyMap["node-c"] != 29 {
		t.Fatalf("expected decoded node latency, got %#v", node.Node.NodeLatencyMap)
	}
	if node.Node.TotalNodes != 4 || node.Node.AliveNodes != 4 {
		t.Fatalf("expected node tree counts from children, got total=%d alive=%d", node.Node.TotalNodes, node.Node.AliveNodes)
	}
	if node.Node.CpuAllocated != 30 || node.Node.CpuLoad != 11.6 ||
		node.Node.LiveStreams != 6 || node.Node.Sessions != 13 || node.Node.AvgDelay != 43 {
		t.Fatalf("expected node runtime fields aggregated from instances and streams, got %#v", node.Node)
	}

	instances, err := svc.ListDashboardInstances(ctx, ListDashboardInstancesInput{
		NodeId: "node-a",
	})
	if err != nil {
		t.Fatalf("list instances: %v", err)
	}
	if instances.NodeInfo == nil || instances.NodeInfo.NodeId != "node-a" {
		t.Fatalf("expected node info for node-a, got %#v", instances.NodeInfo)
	}
	assertDashboardInstanceIDs(t, instances.List, "inst-a", "inst-b")

	runningInstances, err := svc.ListDashboardInstances(ctx, ListDashboardInstancesInput{
		NodeId:  "node-a",
		Status:  "running",
		Keyword: "Transcoder",
	})
	if err != nil {
		t.Fatalf("list filtered instances: %v", err)
	}
	assertDashboardInstanceIDs(t, runningInstances.List, "inst-a")

	childInstances, err := svc.ListDashboardInstances(ctx, ListDashboardInstancesInput{
		NodeId: "node-b",
	})
	if err != nil {
		t.Fatalf("list child node instances: %v", err)
	}
	assertDashboardInstanceIDs(t, childInstances.List, "inst-c")

	keywordInstances, err := svc.ListDashboardInstances(ctx, ListDashboardInstancesInput{
		Keyword: "Node C",
	})
	if err != nil {
		t.Fatalf("list keyword instances: %v", err)
	}
	assertDashboardInstanceIDs(t, keywordInstances.List, "inst-d")

	streams, err := svc.ListDashboardStreams(ctx, ListDashboardStreamsInput{
		SourceType: "node",
		SourceId:   "node-a",
		TenantId:   "tenant-a",
	})
	if err != nil {
		t.Fatalf("list streams: %v", err)
	}
	assertDashboardStreamIDs(t, streams.List, "stream-a", "stream-b")
	if len(streams.List[0].ProtocolSummary) != 2 || streams.List[0].ProtocolSummary[0].ProtocolType != "HLS" {
		t.Fatalf("expected decoded protocol summary, got %#v", streams.List[0].ProtocolSummary)
	}
	if streams.List[0].Duration != 1200 || streams.List[0].ProtocolCount != 2 ||
		streams.List[0].CurrentActiveSessions != 3 || streams.List[0].TotalSessionsLifetime != 30 {
		t.Fatalf("expected stream computed fields from start time, sessions and protocol summary, got %#v", streams.List[0])
	}
	if streams.List[0].ProtocolSummary[0].CurrentSessions != 2 || streams.List[0].ProtocolSummary[1].CurrentSessions != 1 {
		t.Fatalf("expected protocol current sessions from active session rows, got %#v", streams.List[0].ProtocolSummary)
	}
	instanceStreams, err := svc.ListDashboardStreams(ctx, ListDashboardStreamsInput{
		SourceType: "instance",
		SourceId:   "inst-c",
		TenantId:   "tenant-b",
	})
	if err != nil {
		t.Fatalf("list instance streams: %v", err)
	}
	assertDashboardStreamIDs(t, instanceStreams.List, "stream-c")

	pausedStreams, err := svc.ListDashboardStreams(ctx, ListDashboardStreamsInput{
		Status:  "paused",
		Keyword: "backup",
	})
	if err != nil {
		t.Fatalf("list paused streams: %v", err)
	}
	assertDashboardStreamIDs(t, pausedStreams.List, "stream-b")

	sessions, err := svc.ListDashboardSessions(ctx, ListDashboardSessionsInput{
		StreamId: "stream-a",
		TenantId: "tenant-a",
	})
	if err != nil {
		t.Fatalf("list sessions: %v", err)
	}
	if sessions.StreamInfo == nil || sessions.StreamInfo.StreamName != "Camera A" {
		t.Fatalf("expected stream info, got %#v", sessions.StreamInfo)
	}
	if len(sessions.Protocols) != 2 {
		t.Fatalf("expected HLS and RTMP protocol groups, got %#v", sessions.Protocols)
	}
	assertDashboardProtocolGroup(t, sessions.Protocols[0], "HLS", 2, "session-a", "session-b")
	assertDashboardProtocolGroup(t, sessions.Protocols[1], "RTMP", 1, "session-c")

	rtmpSessions, err := svc.ListDashboardSessions(ctx, ListDashboardSessionsInput{
		StreamId:     "stream-a",
		TenantId:     "tenant-a",
		ProtocolType: "RTMP",
	})
	if err != nil {
		t.Fatalf("list rtmp sessions: %v", err)
	}
	if len(rtmpSessions.Protocols) != 1 {
		t.Fatalf("expected one RTMP group, got %#v", rtmpSessions.Protocols)
	}
	assertDashboardProtocolGroup(t, rtmpSessions.Protocols[0], "RTMP", 1, "session-c")

	keywordSessions, err := svc.ListDashboardSessions(ctx, ListDashboardSessionsInput{
		StreamId: "stream-a",
		TenantId: "tenant-a",
		Keyword:  "viewer-b",
	})
	if err != nil {
		t.Fatalf("list keyword sessions: %v", err)
	}
	if len(keywordSessions.Protocols) != 2 {
		t.Fatalf("expected summary protocols to be preserved under keyword filter, got %#v", keywordSessions.Protocols)
	}
	assertDashboardProtocolGroup(t, keywordSessions.Protocols[0], "HLS", 1, "session-b")
	assertDashboardProtocolGroup(t, keywordSessions.Protocols[1], "RTMP", 0)

	if len(sessions.Protocols[0].Sessions[0].LinkHops) != 1 || sessions.Protocols[0].Sessions[0].LinkHops[0].HopIndex != 1 {
		t.Fatalf("expected decoded link hops, got %#v", sessions.Protocols[0].Sessions[0].LinkHops)
	}
	if sessions.Protocols[0].Sessions[0].PlayDuration != 1200 ||
		sessions.Protocols[0].Sessions[0].TotalLinkLatency != 12 {
		t.Fatalf("expected session computed duration and link latency, got %#v", sessions.Protocols[0].Sessions[0])
	}
}

// TestDashboardListQueriesUseFixedUpperBound verifies unpaged dashboard lists are still bounded.
func TestDashboardListQueriesUseFixedUpperBound(t *testing.T) {
	ctx := context.Background()
	setupMediaStrategySQLite(t, ctx)
	setupMediaDashboardReportTables(t, ctx)
	insertDashboardReportNodes(t, ctx)
	insertDashboardInstanceBurst(t, ctx, dashboardReadLimit+5)

	svc := newTestMediaService(t)
	instances, err := svc.ListDashboardInstances(ctx, ListDashboardInstancesInput{
		NodeId: "node-a",
		Status: "running",
	})
	if err != nil {
		t.Fatalf("list bounded instances: %v", err)
	}
	if len(instances.List) != dashboardReadLimit {
		t.Fatalf("expected %d bounded instances, got %d", dashboardReadLimit, len(instances.List))
	}
	if instances.List[0].InstanceId != "bulk-00000" || instances.List[len(instances.List)-1].InstanceId != "bulk-09999" {
		t.Fatalf("unexpected bounded ordering, first=%q last=%q", instances.List[0].InstanceId, instances.List[len(instances.List)-1].InstanceId)
	}
}

// TestBuildDashboardStreamItemUsesActiveSessionCounts verifies stream counters are recalculated.
func TestBuildDashboardStreamItemUsesActiveSessionCounts(t *testing.T) {
	item := buildDashboardStreamItemWithCounts(&dashboardStreamEntity{
		StreamId:              "stream-a",
		ProtocolCount:         2,
		TotalSessionsLifetime: 30,
		CurrentActiveSessions: 999,
		ProtocolSummary:       `[{"protocol_type":"HLS","total_sessions":20,"current_sessions":999},{"protocol_type":"RTMP","total_sessions":10,"current_sessions":999}]`,
	}, map[string]int{
		"HLS":  2,
		"RTMP": 1,
	})
	if item.ProtocolCount != 2 || item.TotalSessionsLifetime != 30 || item.CurrentActiveSessions != 3 {
		t.Fatalf("expected stream counters from protocol summary and active sessions, got %#v", item)
	}
	if item.ProtocolSummary[0].CurrentSessions != 2 || item.ProtocolSummary[1].CurrentSessions != 1 {
		t.Fatalf("expected active session counts merged into protocol summary, got %#v", item.ProtocolSummary)
	}
}

// TestBuildDashboardStreamItemKeepsFallbackLifetime verifies current counts do not become lifetime totals.
func TestBuildDashboardStreamItemKeepsFallbackLifetime(t *testing.T) {
	item := buildDashboardStreamItemWithCounts(&dashboardStreamEntity{
		StreamId:              "stream-a",
		ProtocolCount:         1,
		TotalSessionsLifetime: 42,
		CurrentActiveSessions: 999,
		ProtocolSummary:       `[]`,
	}, map[string]int{"HLS": 3})
	if item.ProtocolCount != 1 || item.TotalSessionsLifetime != 42 || item.CurrentActiveSessions != 3 {
		t.Fatalf("expected fallback lifetime and active session count, got %#v", item)
	}
	if len(item.ProtocolSummary) != 1 || item.ProtocolSummary[0].TotalSessions != 0 || item.ProtocolSummary[0].CurrentSessions != 3 {
		t.Fatalf("expected extra protocol to carry current count only, got %#v", item.ProtocolSummary)
	}
}

func setupMediaDashboardReportTables(t *testing.T, ctx context.Context) {
	t.Helper()

	statements := []string{
		`CREATE TABLE IF NOT EXISTS media_report_node (
			node_id TEXT PRIMARY KEY,
			node_name TEXT NOT NULL DEFAULT '',
			region TEXT NOT NULL DEFAULT '',
			parent_node_id TEXT NOT NULL DEFAULT '0',
			status TEXT NOT NULL DEFAULT '',
			total_nodes INTEGER NOT NULL DEFAULT 0,
			alive_nodes INTEGER NOT NULL DEFAULT 0,
			cpu_allocated REAL NOT NULL DEFAULT 0,
			cpu_load REAL NOT NULL DEFAULT 0,
			memory_allocated REAL NOT NULL DEFAULT 0,
			memory_used REAL NOT NULL DEFAULT 0,
			disk_io_read REAL NOT NULL DEFAULT 0,
			disk_io_write REAL NOT NULL DEFAULT 0,
			network_in REAL NOT NULL DEFAULT 0,
			network_out REAL NOT NULL DEFAULT 0,
			live_streams INTEGER NOT NULL DEFAULT 0,
			sessions INTEGER NOT NULL DEFAULT 0,
			avg_delay INTEGER NOT NULL DEFAULT 0,
			last_heartbeat TEXT,
			node_latency_map TEXT NOT NULL DEFAULT '{}',
			report_time INTEGER NOT NULL,
			updated_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS media_report_instance (
			instance_id TEXT PRIMARY KEY,
			instance_name TEXT NOT NULL DEFAULT '',
			node_id TEXT NOT NULL DEFAULT '',
			node_name TEXT NOT NULL DEFAULT '',
			region TEXT NOT NULL DEFAULT '',
			node_status TEXT NOT NULL DEFAULT '',
			status TEXT NOT NULL DEFAULT '',
			cpu_allocated REAL NOT NULL DEFAULT 0,
			cpu_load REAL NOT NULL DEFAULT 0,
			memory_allocated REAL NOT NULL DEFAULT 0,
			memory_used REAL NOT NULL DEFAULT 0,
			disk_io_read REAL NOT NULL DEFAULT 0,
			disk_io_write REAL NOT NULL DEFAULT 0,
			network_in REAL NOT NULL DEFAULT 0,
			network_out REAL NOT NULL DEFAULT 0,
			live_streams INTEGER NOT NULL DEFAULT 0,
			sessions INTEGER NOT NULL DEFAULT 0,
			start_time TEXT,
			version TEXT NOT NULL DEFAULT '',
			report_time INTEGER NOT NULL,
			updated_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS media_report_stream (
			stream_id TEXT PRIMARY KEY,
			source_type TEXT NOT NULL DEFAULT '',
			source_id TEXT NOT NULL DEFAULT '',
			tenant_id TEXT NOT NULL DEFAULT '',
			node_id TEXT NOT NULL DEFAULT '',
			node_name TEXT NOT NULL DEFAULT '',
			instance_id TEXT NOT NULL DEFAULT '',
			instance_name TEXT NOT NULL DEFAULT '',
			source_url TEXT NOT NULL DEFAULT '',
			stream_name TEXT NOT NULL DEFAULT '',
			resolution TEXT NOT NULL DEFAULT '',
			fps REAL NOT NULL DEFAULT 0,
			bitrate INTEGER NOT NULL DEFAULT 0,
			packet_loss REAL NOT NULL DEFAULT 0,
			status TEXT NOT NULL DEFAULT '',
			start_time TEXT,
			duration INTEGER NOT NULL DEFAULT 0,
			avg_delay INTEGER NOT NULL DEFAULT 0,
			protocol_count INTEGER NOT NULL DEFAULT 0,
			total_sessions_lifetime INTEGER NOT NULL DEFAULT 0,
			current_active_sessions INTEGER NOT NULL DEFAULT 0,
			watermark_enabled INTEGER NOT NULL DEFAULT 0,
			protocol_summary TEXT NOT NULL DEFAULT '[]',
			report_time INTEGER NOT NULL,
			updated_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS media_report_session (
			session_id TEXT PRIMARY KEY,
			stream_id TEXT NOT NULL DEFAULT '',
			stream_name TEXT NOT NULL DEFAULT '',
			tenant_id TEXT NOT NULL DEFAULT '',
			client_id TEXT NOT NULL DEFAULT '',
			client_ip TEXT,
			client_type TEXT NOT NULL DEFAULT '',
			user_name TEXT NOT NULL DEFAULT '',
			protocol_type TEXT NOT NULL DEFAULT '',
			start_time TEXT,
			play_duration INTEGER NOT NULL DEFAULT 0,
			current_fps REAL NOT NULL DEFAULT 0,
			current_bitrate INTEGER NOT NULL DEFAULT 0,
			current_resolution TEXT NOT NULL DEFAULT '',
			node_id TEXT NOT NULL DEFAULT '',
			node_name TEXT NOT NULL DEFAULT '',
			instance_id TEXT NOT NULL DEFAULT '',
			instance_name TEXT NOT NULL DEFAULT '',
			link_hops TEXT NOT NULL DEFAULT '[]',
			total_link_latency INTEGER NOT NULL DEFAULT 0,
			report_time INTEGER NOT NULL,
			updated_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP
		)`,
	}
	for _, statement := range statements {
		if _, err := dao.MediaReportNode.DB().Exec(ctx, statement); err != nil {
			t.Fatalf("exec sqlite schema: %v", err)
		}
	}

	cleanupStatements := []string{
		`DELETE FROM media_report_session`,
		`DELETE FROM media_report_stream`,
		`DELETE FROM media_report_instance`,
		`DELETE FROM media_report_node`,
	}
	for _, statement := range cleanupStatements {
		if _, err := dao.MediaReportNode.DB().Exec(ctx, statement); err != nil {
			t.Fatalf("cleanup dashboard sqlite data: %v", err)
		}
	}
}

func insertDashboardReportFixtures(t *testing.T, ctx context.Context) {
	t.Helper()

	reportTime := int64(1780000000000)
	baseTime := gtime.NewFromTime(time.UnixMilli(reportTime - 1_200_000))
	insertDashboardReportNodes(t, ctx)
	insertDashboardReports(t, ctx, []any{
		do.MediaReportInstance{
			InstanceId:      "inst-a",
			InstanceName:    "Transcoder A",
			NodeId:          "node-a",
			NodeName:        "Node A",
			Region:          "East",
			NodeStatus:      "healthy",
			Status:          "running",
			CpuAllocated:    8,
			CpuLoad:         3.5,
			MemoryAllocated: 16,
			MemoryUsed:      7,
			LiveStreams:     2,
			Sessions:        4,
			StartTime:       baseTime,
			Version:         "v1.0.0",
			ReportTime:      reportTime,
		},
		do.MediaReportInstance{
			InstanceId:      "inst-b",
			InstanceName:    "Recorder B",
			NodeId:          "node-a",
			NodeName:        "Node A",
			Region:          "East",
			NodeStatus:      "healthy",
			Status:          "stopped",
			CpuAllocated:    4,
			CpuLoad:         0.2,
			MemoryAllocated: 8,
			MemoryUsed:      1.1,
			LiveStreams:     0,
			Sessions:        0,
			StartTime:       baseTime,
			Version:         "v1.0.1",
			ReportTime:      reportTime - 1,
		},
		do.MediaReportInstance{
			InstanceId:      "inst-c",
			InstanceName:    "Relay C",
			NodeId:          "node-b",
			NodeName:        "Node B",
			Region:          "East",
			NodeStatus:      "healthy",
			Status:          "running",
			CpuAllocated:    12,
			CpuLoad:         5.5,
			MemoryAllocated: 32,
			MemoryUsed:      18,
			LiveStreams:     3,
			Sessions:        7,
			StartTime:       baseTime,
			Version:         "v1.1.0",
			ReportTime:      reportTime,
		},
		do.MediaReportInstance{
			InstanceId:      "inst-d",
			InstanceName:    "Node C Edge",
			NodeId:          "node-c",
			NodeName:        "Node C",
			Region:          "West",
			NodeStatus:      "degraded",
			Status:          "running",
			CpuAllocated:    6,
			CpuLoad:         2.4,
			MemoryAllocated: 12,
			MemoryUsed:      4,
			LiveStreams:     1,
			Sessions:        2,
			StartTime:       baseTime,
			Version:         "v1.2.0",
			ReportTime:      reportTime,
		},
	})
	insertDashboardReports(t, ctx, []any{
		do.MediaReportStream{
			StreamId:              "stream-a",
			SourceType:            "node",
			SourceId:              "node-a",
			TenantId:              "tenant-a",
			NodeId:                "node-a",
			NodeName:              "Node A",
			InstanceId:            "inst-a",
			InstanceName:          "Transcoder A",
			SourceUrl:             "https://example.test/stream-a.flv",
			StreamName:            "Camera A",
			Resolution:            "1920x1080",
			Fps:                   25,
			Bitrate:               4000,
			PacketLoss:            0.12,
			Status:                "playing",
			StartTime:             baseTime,
			Duration:              0,
			AvgDelay:              40,
			ProtocolCount:         2,
			TotalSessionsLifetime: 30,
			CurrentActiveSessions: 999,
			WatermarkEnabled:      true,
			ProtocolSummary:       `[{"protocol_type":"HLS","total_sessions":20,"current_sessions":2},{"protocol_type":"RTMP","total_sessions":10,"current_sessions":1}]`,
			ReportTime:            reportTime,
		},
		do.MediaReportStream{
			StreamId:              "stream-b",
			SourceType:            "node",
			SourceId:              "node-a",
			TenantId:              "tenant-a",
			NodeId:                "node-a",
			NodeName:              "Node A",
			InstanceId:            "inst-b",
			InstanceName:          "Recorder B",
			SourceUrl:             "https://example.test/stream-b.flv",
			StreamName:            "Backup Camera B",
			Resolution:            "1280x720",
			Fps:                   20,
			Bitrate:               2200,
			PacketLoss:            0.3,
			Status:                "paused",
			StartTime:             baseTime,
			Duration:              120,
			AvgDelay:              55,
			ProtocolCount:         1,
			TotalSessionsLifetime: 4,
			CurrentActiveSessions: 0,
			WatermarkEnabled:      false,
			ProtocolSummary:       `[{"protocol_type":"HLS","total_sessions":4,"current_sessions":0}]`,
			ReportTime:            reportTime - 1,
		},
		do.MediaReportStream{
			StreamId:              "stream-c",
			SourceType:            "instance",
			SourceId:              "inst-c",
			TenantId:              "tenant-b",
			NodeId:                "node-b",
			NodeName:              "Node B",
			InstanceId:            "inst-c",
			InstanceName:          "Relay C",
			SourceUrl:             "https://example.test/stream-c.flv",
			StreamName:            "Camera C",
			Resolution:            "1920x1080",
			Fps:                   25,
			Bitrate:               3500,
			PacketLoss:            0.05,
			Status:                "playing",
			StartTime:             baseTime,
			Duration:              600,
			AvgDelay:              35,
			ProtocolCount:         1,
			TotalSessionsLifetime: 7,
			CurrentActiveSessions: 2,
			WatermarkEnabled:      true,
			ProtocolSummary:       `[{"protocol_type":"FLV","total_sessions":7,"current_sessions":2}]`,
			ReportTime:            reportTime,
		},
	})
	insertDashboardReports(t, ctx, []any{
		dashboardSessionDO("session-a", "stream-a", "Camera A", "tenant-a", "client-a", "viewer-a", "HLS", "inst-a", 0, reportTime, baseTime),
		dashboardSessionDO("session-b", "stream-a", "Camera A", "tenant-a", "client-b", "viewer-b", "HLS", "inst-a", 100, reportTime-1, baseTime),
		dashboardSessionDO("session-c", "stream-a", "Camera A", "tenant-a", "client-c", "viewer-c", "RTMP", "inst-a", 80, reportTime-2, baseTime),
		dashboardSessionDO("session-d", "stream-c", "Camera C", "tenant-b", "client-d", "viewer-d", "FLV", "inst-c", 60, reportTime, baseTime),
	})
}

func insertDashboardReportNodes(t *testing.T, ctx context.Context) {
	t.Helper()

	baseTime := gtime.NewFromTime(time.Date(2026, 6, 15, 8, 0, 0, 0, time.UTC))
	reportTime := int64(1780000000000)
	insertDashboardReports(t, ctx, []any{
		do.MediaReportNode{
			NodeId:          "node-a",
			NodeName:        "Node A",
			Region:          "East",
			ParentNodeId:    dashboardRootNodeID,
			Status:          "healthy",
			TotalNodes:      4,
			AliveNodes:      3,
			CpuAllocated:    64,
			CpuLoad:         32.5,
			MemoryAllocated: 128,
			MemoryUsed:      64,
			NetworkIn:       120,
			NetworkOut:      240,
			LiveStreams:     8,
			Sessions:        16,
			AvgDelay:        30,
			LastHeartbeat:   baseTime,
			NodeLatencyMap:  `{"node-b":17,"node-c":29}`,
			ReportTime:      reportTime,
		},
		do.MediaReportNode{
			NodeId:         "node-b",
			NodeName:       "Node B",
			Region:         "East",
			ParentNodeId:   "node-a",
			Status:         "healthy",
			NodeLatencyMap: `{"node-d":8}`,
			LastHeartbeat:  baseTime,
			ReportTime:     reportTime,
		},
		do.MediaReportNode{
			NodeId:         "node-c",
			NodeName:       "Node C",
			Region:         "West",
			ParentNodeId:   "node-a",
			Status:         "degraded",
			NodeLatencyMap: `{}`,
			LastHeartbeat:  baseTime,
			ReportTime:     reportTime - 1,
		},
		do.MediaReportNode{
			NodeId:         "node-d",
			NodeName:       "Node D",
			Region:         "East",
			ParentNodeId:   "node-b",
			Status:         "healthy",
			NodeLatencyMap: `{}`,
			LastHeartbeat:  baseTime,
			ReportTime:     reportTime - 2,
		},
	})
}

func insertDashboardInstanceBurst(t *testing.T, ctx context.Context, count int) {
	t.Helper()

	baseTime := gtime.NewFromTime(time.Date(2026, 6, 15, 8, 0, 0, 0, time.UTC))
	err := dao.MediaReportInstance.Transaction(ctx, func(ctx context.Context, _ gdb.TX) error {
		for i := 0; i < count; i++ {
			if _, err := dao.MediaReportInstance.Ctx(ctx).Data(do.MediaReportInstance{
				InstanceId:      fmt.Sprintf("bulk-%05d", i),
				InstanceName:    fmt.Sprintf("Bulk Instance %05d", i),
				NodeId:          "node-a",
				NodeName:        "Node A",
				Region:          "East",
				NodeStatus:      "healthy",
				Status:          "running",
				CpuAllocated:    2,
				CpuLoad:         1,
				MemoryAllocated: 4,
				MemoryUsed:      2,
				StartTime:       baseTime,
				Version:         "v-bulk",
				ReportTime:      int64(1780000000000),
			}).Insert(); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		t.Fatalf("insert dashboard instance burst: %v", err)
	}
}

func insertDashboardReports(t *testing.T, ctx context.Context, rows []any) {
	t.Helper()

	for _, row := range rows {
		switch value := row.(type) {
		case do.MediaReportNode:
			if _, err := dao.MediaReportNode.Ctx(ctx).Data(value).Insert(); err != nil {
				t.Fatalf("insert node report: %v", err)
			}
		case do.MediaReportInstance:
			if _, err := dao.MediaReportInstance.Ctx(ctx).Data(value).Insert(); err != nil {
				t.Fatalf("insert instance report: %v", err)
			}
		case do.MediaReportStream:
			if _, err := dao.MediaReportStream.Ctx(ctx).Data(value).Insert(); err != nil {
				t.Fatalf("insert stream report: %v", err)
			}
		case do.MediaReportSession:
			if _, err := dao.MediaReportSession.Ctx(ctx).Data(value).Insert(); err != nil {
				t.Fatalf("insert session report: %v", err)
			}
		default:
			t.Fatalf("unsupported dashboard report row %#v", row)
		}
	}
}

func dashboardSessionDO(
	sessionID string,
	streamID string,
	streamName string,
	tenantID string,
	clientID string,
	userName string,
	protocolType string,
	instanceID string,
	playDuration int,
	reportTime int64,
	startTime *gtime.Time,
) do.MediaReportSession {
	return do.MediaReportSession{
		SessionId:         sessionID,
		StreamId:          streamID,
		StreamName:        streamName,
		TenantId:          tenantID,
		ClientId:          clientID,
		ClientIp:          "192.0.2.10",
		ClientType:        "web",
		UserName:          userName,
		ProtocolType:      protocolType,
		StartTime:         startTime,
		PlayDuration:      playDuration,
		CurrentFps:        25,
		CurrentBitrate:    2500,
		CurrentResolution: "1280x720",
		NodeId:            "node-a",
		NodeName:          "Node A",
		InstanceId:        instanceID,
		InstanceName:      "Transcoder A",
		LinkHops:          `[{"hop_index":1,"node_id":"node-a","latency_ms":12}]`,
		TotalLinkLatency:  999,
		ReportTime:        reportTime,
	}
}

func assertDashboardInstanceIDs(t *testing.T, items []*DashboardInstanceItem, ids ...string) {
	t.Helper()

	if len(items) != len(ids) {
		t.Fatalf("expected instance ids %v, got %#v", ids, dashboardInstanceIDs(items))
	}
	for i, id := range ids {
		if items[i].InstanceId != id {
			t.Fatalf("expected instance ids %v, got %#v", ids, dashboardInstanceIDs(items))
		}
	}
}

func dashboardInstanceIDs(items []*DashboardInstanceItem) []string {
	ids := make([]string, 0, len(items))
	for _, item := range items {
		if item == nil {
			ids = append(ids, "")
			continue
		}
		ids = append(ids, item.InstanceId)
	}
	return ids
}

func assertDashboardStreamIDs(t *testing.T, items []*DashboardStreamItem, ids ...string) {
	t.Helper()

	if len(items) != len(ids) {
		t.Fatalf("expected stream ids %v, got %#v", ids, dashboardStreamIDs(items))
	}
	for i, id := range ids {
		if items[i].StreamId != id {
			t.Fatalf("expected stream ids %v, got %#v", ids, dashboardStreamIDs(items))
		}
	}
}

func dashboardStreamIDs(items []*DashboardStreamItem) []string {
	ids := make([]string, 0, len(items))
	for _, item := range items {
		if item == nil {
			ids = append(ids, "")
			continue
		}
		ids = append(ids, item.StreamId)
	}
	return ids
}

func assertDashboardProtocolGroup(t *testing.T, protocol *DashboardSessionProtocol, protocolType string, sessionCount int, sessionIDs ...string) {
	t.Helper()

	if protocol == nil {
		t.Fatalf("expected protocol %s, got nil", protocolType)
	}
	if protocol.ProtocolType != protocolType || protocol.SessionCount != sessionCount {
		t.Fatalf("expected protocol %s count %d, got %#v", protocolType, sessionCount, protocol)
	}
	if len(protocol.Sessions) != len(sessionIDs) {
		t.Fatalf("expected sessions %v, got %#v", sessionIDs, dashboardSessionIDs(protocol.Sessions))
	}
	for i, id := range sessionIDs {
		if protocol.Sessions[i].SessionId != id {
			t.Fatalf("expected sessions %v, got %#v", sessionIDs, dashboardSessionIDs(protocol.Sessions))
		}
	}
}

func dashboardSessionIDs(items []*DashboardSessionItem) []string {
	ids := make([]string, 0, len(items))
	for _, item := range items {
		if item == nil {
			ids = append(ids, "")
			continue
		}
		ids = append(ids, item.SessionId)
	}
	return ids
}
