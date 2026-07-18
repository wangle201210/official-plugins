// This file verifies dashboard controller methods preserve the frontend DTO shape.

package media

import (
	"context"
	"encoding/json"
	"testing"

	"lina-plugin-media/backend/api/media/v1"
	mediasvc "lina-plugin-media/backend/internal/service/media"
)

// TestDashboardControllerReturnsFrontendDataShape verifies dashboard controller projection with multi-row mocks.
func TestDashboardControllerReturnsFrontendDataShape(t *testing.T) {
	ctx := context.Background()
	controller := &ControllerV1{mediaSvc: &dashboardControllerService{}}

	node, err := controller.GetDashboardNodeOverview(ctx, &v1.GetDashboardNodeOverviewReq{RootNodeId: "node-a"})
	if err != nil {
		t.Fatalf("get dashboard node overview: %v", err)
	}
	nodePayload := mustMarshalDashboardControllerShape(t, node)
	assertDashboardControllerExactKeys(t, nodePayload,
		"node_id", "node_name", "region", "status", "parent_node_id", "total_nodes", "alive_nodes",
		"cpu_allocated", "cpu_load", "memory_allocated", "memory_used", "disk_io_read", "disk_io_write",
		"network_in", "network_out", "live_streams", "sessions", "avg_delay", "last_heartbeat",
		"report_time", "node_latency_map", "child_nodes",
	)
	assertDashboardControllerMissingKey(t, nodePayload, "total")
	if len(node.ChildNodes) != 2 {
		t.Fatalf("expected two child nodes, got %#v", node.ChildNodes)
	}

	instances, err := controller.ListDashboardInstances(ctx, &v1.ListDashboardInstancesReq{NodeId: "node-a"})
	if err != nil {
		t.Fatalf("list dashboard instances: %v", err)
	}
	instancePayload := mustMarshalDashboardControllerShape(t, instances)
	assertDashboardControllerExactKeys(t, instancePayload, "node_info", "instance_list")
	assertDashboardControllerMissingKey(t, instancePayload, "total")
	if instances.NodeInfo == nil || instances.NodeInfo.NodeId != "node-a" || len(instances.InstanceList) != 2 {
		t.Fatalf("unexpected instance response %#v", instances)
	}

	streams, err := controller.ListDashboardStreams(ctx, &v1.ListDashboardStreamsReq{TenantId: "tenant-a"})
	if err != nil {
		t.Fatalf("list dashboard streams: %v", err)
	}
	streamPayload := mustMarshalDashboardControllerShape(t, streams)
	assertDashboardControllerExactKeys(t, streamPayload, "source_type", "source_id", "stream_list")
	assertDashboardControllerMissingKey(t, streamPayload, "total")
	if len(streams.StreamList) != 2 || len(streams.StreamList[0].ProtocolSummary) != 2 {
		t.Fatalf("unexpected stream response %#v", streams)
	}
	if streams.StreamList[0].DeviceId != "device-a" {
		t.Fatalf("expected stream device id, got %#v", streams.StreamList[0])
	}

	sessions, err := controller.ListDashboardSessions(ctx, &v1.ListDashboardSessionsReq{StreamId: "stream-a"})
	if err != nil {
		t.Fatalf("list dashboard sessions: %v", err)
	}
	sessionPayload := mustMarshalDashboardControllerShape(t, sessions)
	assertDashboardControllerExactKeys(t, sessionPayload, "stream_info", "protocol_list")
	assertDashboardControllerMissingKey(t, sessionPayload, "total")
	if len(sessions.ProtocolList) != 2 || len(sessions.ProtocolList[0].ActiveSessionList) != 2 {
		t.Fatalf("unexpected session response %#v", sessions)
	}
	if sessions.StreamInfo.DeviceId != "device-a" || sessions.ProtocolList[0].ActiveSessionList[0].DeviceId != "device-a" {
		t.Fatalf("expected session device id projection, got %#v", sessions)
	}

	topology, err := controller.GetDashboardTopology(ctx, &v1.GetDashboardTopologyReq{DeviceId: "device-a"})
	if err != nil {
		t.Fatalf("get dashboard topology: %v", err)
	}
	topologyPayload := mustMarshalDashboardControllerShape(t, topology)
	assertDashboardControllerExactKeys(t, topologyPayload,
		"node_id", "node_name", "device_id", "stream_count", "session_count",
		"session_detail_limited", "generated_at", "devices",
	)
	assertDashboardControllerMissingKey(t, topologyPayload, "total")
	if len(topology.Devices) != 1 || len(topology.Devices[0].Streams) != 1 {
		t.Fatalf("unexpected topology response %#v", topology)
	}
	if topology.Devices[0].Streams[0].Gateway.InstanceId != "inst-a" ||
		topology.Devices[0].Streams[0].Protocols[0].Tenants[0].Users[0].ClientIp != "192.0.2.11" {
		t.Fatalf("expected topology fields backed by report projections, got %#v", topology.Devices[0].Streams[0])
	}
}

type dashboardControllerService struct {
	mediasvc.Service
}

func (s *dashboardControllerService) GetDashboardNodeOverview(context.Context, mediasvc.DashboardNodeOverviewInput) (*mediasvc.DashboardNodeOverviewOutput, error) {
	lastHeartbeat := int64(1780000000000)
	return &mediasvc.DashboardNodeOverviewOutput{Node: &mediasvc.DashboardNodeOverviewItem{
		NodeId:         "node-a",
		NodeName:       "Node A",
		Region:         "East",
		Status:         "healthy",
		ParentNodeId:   "0",
		TotalNodes:     3,
		AliveNodes:     2,
		LiveStreams:    5,
		Sessions:       11,
		AvgDelay:       28,
		LastHeartbeat:  &lastHeartbeat,
		ReportTime:     1780000000000,
		NodeLatencyMap: map[string]int{"node-b": 12, "node-c": 24},
		ChildNodes: []*mediasvc.DashboardNodeOverviewItem{
			{NodeId: "node-b", NodeName: "Node B", ParentNodeId: "node-a", NodeLatencyMap: map[string]int{}, ChildNodes: []*mediasvc.DashboardNodeOverviewItem{}},
			{NodeId: "node-c", NodeName: "Node C", ParentNodeId: "node-a", NodeLatencyMap: map[string]int{}, ChildNodes: []*mediasvc.DashboardNodeOverviewItem{}},
		},
	}}, nil
}

func (s *dashboardControllerService) ListDashboardInstances(context.Context, mediasvc.ListDashboardInstancesInput) (*mediasvc.ListDashboardInstancesOutput, error) {
	startTime := int64(1780000000000)
	return &mediasvc.ListDashboardInstancesOutput{
		NodeInfo: &mediasvc.DashboardInstanceNodeInfo{NodeId: "node-a", NodeName: "Node A", Region: "East", Status: "healthy"},
		List: []*mediasvc.DashboardInstanceItem{
			{InstanceId: "inst-a", InstanceName: "Transcoder A", Status: "running", StartTime: &startTime, LiveStreams: 3, Sessions: 8, Version: "v1.0.0"},
			{InstanceId: "inst-b", InstanceName: "Recorder B", Status: "stopped", StartTime: &startTime, Version: "v1.0.1"},
		},
	}, nil
}

func (s *dashboardControllerService) GetDashboardTopology(context.Context, mediasvc.DashboardTopologyInput) (*mediasvc.DashboardTopologyOutput, error) {
	startTime := int64(1780000000000)
	return &mediasvc.DashboardTopologyOutput{
		NodeId:       "node-a",
		NodeName:     "Node A",
		DeviceId:     "device-a",
		StreamCount:  1,
		SessionCount: 2,
		GeneratedAt:  1780000000000,
		Devices: []*mediasvc.DashboardTopologyDevice{
			{
				DeviceId:     "device-a",
				StreamCount:  1,
				SessionCount: 2,
				Streams: []*mediasvc.DashboardTopologyStream{
					{
						StreamId:     "stream-a",
						StreamName:   "Camera A",
						Status:       "playing",
						BasePlatform: &mediasvc.DashboardTopologyBasePlatform{Bitrate: 4000, Resolution: "1920x1080", StreamProtocol: "RTSP", DeviceCode: "device-a"},
						Gateway:      &mediasvc.DashboardTopologyGateway{StreamProtocol: "HLS", Resolution: "1920x1080", Bitrate: 4000, Fps: 25, NodeId: "node-a", InstanceId: "inst-a"},
						Protocols: []*mediasvc.DashboardTopologyProtocol{
							{
								ProtocolType: "HLS",
								ReuseCount:   2,
								Tenants: []*mediasvc.DashboardTopologyTenant{
									{
										TenantId:           "tenant-a",
										ConcurrentSessions: 2,
										Users: []*mediasvc.DashboardTopologyUser{
											{SessionId: "session-a", UserName: "viewer-a", ClientId: "client-a", ClientIp: "192.0.2.11", ClientType: 2, ProtocolType: "HLS", TenantId: "tenant-a", NodeId: "node-a", InstanceId: "inst-a", StartTime: &startTime, PlayDuration: 60},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}, nil
}

func (s *dashboardControllerService) ListDashboardStreams(context.Context, mediasvc.ListDashboardStreamsInput) (*mediasvc.ListDashboardStreamsOutput, error) {
	startTime := int64(1780000000000)
	return &mediasvc.ListDashboardStreamsOutput{
		SourceType: "node",
		SourceId:   "node-a",
		List: []*mediasvc.DashboardStreamItem{
			{
				StreamId:              "stream-a",
				StreamName:            "Camera A",
				DeviceId:              "device-a",
				SourceUrl:             "https://example.test/a.flv",
				Status:                "playing",
				StartTime:             &startTime,
				ProtocolCount:         2,
				TotalSessionsLifetime: 30,
				CurrentActiveSessions: 3,
				ProtocolSummary: []*mediasvc.DashboardProtocolItem{
					{ProtocolType: "HLS", TotalSessions: 20, CurrentSessions: 2},
					{ProtocolType: "RTMP", TotalSessions: 10, CurrentSessions: 1},
				},
			},
			{StreamId: "stream-b", StreamName: "Camera B", DeviceId: "device-b", Status: "paused", ProtocolSummary: []*mediasvc.DashboardProtocolItem{}},
		},
	}, nil
}

func (s *dashboardControllerService) ListDashboardSessions(context.Context, mediasvc.ListDashboardSessionsInput) (*mediasvc.ListDashboardSessionsOutput, error) {
	startTime := int64(1780000000000)
	return &mediasvc.ListDashboardSessionsOutput{
		StreamInfo: &mediasvc.DashboardSessionStreamInfo{StreamId: "stream-a", StreamName: "Camera A", DeviceId: "device-a"},
		Protocols: []*mediasvc.DashboardSessionProtocol{
			{
				ProtocolType: "HLS",
				Status:       "active",
				SessionCount: 2,
				Sessions: []*mediasvc.DashboardSessionItem{
					{SessionId: "session-a", DeviceId: "device-a", ClientId: "client-a", TenantId: "tenant-a", ProtocolType: "HLS", StartTime: &startTime, LinkHops: []*mediasvc.DashboardLinkHopItem{{HopIndex: 1, NodeId: "node-a", LatencyMs: 12}}},
					{SessionId: "session-b", DeviceId: "device-a", ClientId: "client-b", TenantId: "tenant-a", ProtocolType: "HLS", StartTime: &startTime, LinkHops: []*mediasvc.DashboardLinkHopItem{{HopIndex: 1, NodeId: "node-b", LatencyMs: 20}}},
				},
			},
			{ProtocolType: "RTMP", Status: "active", SessionCount: 1, Sessions: []*mediasvc.DashboardSessionItem{{SessionId: "session-c", ProtocolType: "RTMP", StartTime: &startTime}}},
		},
	}, nil
}

func mustMarshalDashboardControllerShape(t *testing.T, value any) map[string]any {
	t.Helper()

	raw, err := json.Marshal(value)
	if err != nil {
		t.Fatalf("marshal dashboard controller response: %v", err)
	}
	payload := make(map[string]any)
	if err := json.Unmarshal(raw, &payload); err != nil {
		t.Fatalf("unmarshal dashboard controller response: %v", err)
	}
	return payload
}

func assertDashboardControllerExactKeys(t *testing.T, payload map[string]any, keys ...string) {
	t.Helper()

	if len(payload) != len(keys) {
		t.Fatalf("expected keys %v, got payload %#v", keys, payload)
	}
	for _, key := range keys {
		if _, ok := payload[key]; !ok {
			t.Fatalf("expected key %q in payload %#v", key, payload)
		}
	}
}

func assertDashboardControllerMissingKey(t *testing.T, payload map[string]any, key string) {
	t.Helper()

	if _, ok := payload[key]; ok {
		t.Fatalf("did not expect key %q in payload %#v", key, payload)
	}
}
