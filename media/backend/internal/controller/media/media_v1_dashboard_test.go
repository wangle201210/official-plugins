// This file verifies dashboard controller methods preserve the frontend DTO shape.

package media

import (
	"bytes"
	"context"
	"encoding/json"
	"math"
	"reflect"
	"testing"

	"lina-plugin-media/backend/api/media/v1"
	mediasvc "lina-plugin-media/backend/internal/service/media"
)

// TestDashboardNetworkRatesUseDecimalKilobits exercises every public resource
// projection while preserving stored values, disk units and raw bps reports.
func TestDashboardNetworkRatesUseDecimalKilobits(t *testing.T) {
	for _, test := range []struct {
		name                    string
		storedIn, storedOut     float64
		wantInKbps, wantOutKbps float64
	}{
		{name: "zero"},
		{name: "fractional", storedIn: 0.125, storedOut: 0.25, wantInKbps: 1.024, wantOutKbps: 2.048},
		{name: "eight_megabits", storedIn: 976.5625, storedOut: 488.28125, wantInKbps: 8000, wantOutKbps: 4000},
		{name: "stored_two_decimals", storedIn: 976.56, storedOut: 488.28, wantInKbps: 7999.97952, wantOutKbps: 3999.98976},
	} {
		t.Run(test.name, func(t *testing.T) {
			const diskRead, diskWrite = 7.25, 11.5
			var (
				rawReport = map[string]any{"network_in": json.Number("8000000"), "network_out": json.Number("4000000")}
				child     = &mediasvc.DashboardNodeOverviewItem{
					NodeId: "child", NetworkIn: test.storedOut, NetworkOut: test.storedIn,
					DiskIoRead: diskRead, DiskIoWrite: diskWrite,
					InstanceResources: &mediasvc.DashboardNodeResources{
						NetworkIn: test.storedOut * 2, NetworkOut: test.storedIn * 2, DiskIoRead: diskRead, DiskIoWrite: diskWrite,
					},
				}
				root = &mediasvc.DashboardNodeOverviewItem{
					NodeId: "0", NetworkIn: test.storedIn, NetworkOut: test.storedOut,
					DiskIoRead: diskRead, DiskIoWrite: diskWrite,
					InstanceResources: &mediasvc.DashboardNodeResources{
						NetworkIn: test.storedIn / 2, NetworkOut: test.storedOut / 2, DiskIoRead: diskRead, DiskIoWrite: diskWrite,
					},
					ChildNodes:   []*mediasvc.DashboardNodeOverviewItem{child},
					NetworkLinks: []*mediasvc.DashboardNetworkLink{{ReportPayload: rawReport}},
				}
				instance = &mediasvc.DashboardInstanceItem{
					InstanceId: "instance", NetworkIn: test.storedIn, NetworkOut: test.storedOut,
					DiskIoRead: diskRead, DiskIoWrite: diskWrite, ReportPayload: rawReport,
				}
				nodeDTO     = dashboardNodeOverviewToDTO(root)
				instanceDTO = dashboardInstanceListToDTO([]*mediasvc.DashboardInstanceItem{instance})[0]
			)
			for _, projection := range []struct {
				name                                  string
				in, out, read, write, wantIn, wantOut float64
			}{
				{"root", nodeDTO.NetworkIn, nodeDTO.NetworkOut, nodeDTO.DiskIoRead, nodeDTO.DiskIoWrite, test.wantInKbps, test.wantOutKbps},
				{"child", nodeDTO.ChildNodes[0].NetworkIn, nodeDTO.ChildNodes[0].NetworkOut, nodeDTO.ChildNodes[0].DiskIoRead, nodeDTO.ChildNodes[0].DiskIoWrite, test.wantOutKbps, test.wantInKbps},
				{"root_resources", nodeDTO.InstanceResources.NetworkIn, nodeDTO.InstanceResources.NetworkOut, nodeDTO.InstanceResources.DiskIoRead, nodeDTO.InstanceResources.DiskIoWrite, test.wantInKbps / 2, test.wantOutKbps / 2},
				{"child_resources", nodeDTO.ChildNodes[0].InstanceResources.NetworkIn, nodeDTO.ChildNodes[0].InstanceResources.NetworkOut, nodeDTO.ChildNodes[0].InstanceResources.DiskIoRead, nodeDTO.ChildNodes[0].InstanceResources.DiskIoWrite, test.wantOutKbps * 2, test.wantInKbps * 2},
				{"instance_list", instanceDTO.NetworkIn, instanceDTO.NetworkOut, instanceDTO.DiskIoRead, instanceDTO.DiskIoWrite, test.wantInKbps, test.wantOutKbps},
			} {
				if math.Abs(projection.in-projection.wantIn) > 1e-9 || math.Abs(projection.out-projection.wantOut) > 1e-9 {
					t.Fatalf("%s network rate must use decimal Kbps: got=(%v,%v) want=(%v,%v)", projection.name, projection.in, projection.out, projection.wantIn, projection.wantOut)
				}
				if projection.read != diskRead || projection.write != diskWrite {
					t.Fatalf("%s disk IO units changed during network conversion", projection.name)
				}
			}
			for _, payload := range []map[string]any{nodeDTO.NetworkLinks[0].ReportPayload, instanceDTO.ReportPayload} {
				if payload["network_in"] != json.Number("8000000") || payload["network_out"] != json.Number("4000000") {
					t.Fatalf("raw report must retain its original bps values: %#v", payload)
				}
			}
			if root.NetworkIn != test.storedIn || root.NetworkOut != test.storedOut || child.NetworkIn != test.storedOut ||
				instance.NetworkIn != test.storedIn || instance.NetworkOut != test.storedOut {
				t.Fatal("DTO mapping mutated service values or would convert them twice on a later read")
			}
		})
	}
}

// TestDashboardControllerReturnsFrontendDataShape verifies dashboard controller projection with multi-row mocks.
func TestDashboardControllerReturnsFrontendDataShape(t *testing.T) {
	ctx := t.Context()
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
		"report_time", "node_latency_map", "network_links", "network_links_limited", "instance_resources", "child_nodes",
	)
	assertDashboardControllerMissingKey(t, nodePayload, "total")
	if len(node.ChildNodes) != 2 {
		t.Fatalf("expected two child nodes, got %#v", node.ChildNodes)
	}
	if len(node.NetworkLinks) != 1 || node.NetworkLinks[0].PacketLoss != 0.02 || node.NetworkLinks[0].Jitter != 3 ||
		node.NetworkLinks[0].Reachable != nil || node.NetworkLinks[0].ReportTime != 1791784485000 {
		t.Fatalf("expected complete raw network metrics: %#v", node.NetworkLinks)
	}
	if node.InstanceResources == nil || node.InstanceResources.InstanceCount != 2 || node.InstanceResources.CpuLoad != 2.5 {
		t.Fatalf("expected direct instance resource projection: %#v", node.InstanceResources)
	}
	if node.ChildNodes[0].NetworkLinks == nil || node.ChildNodes[0].InstanceResources == nil {
		t.Fatal("empty child metrics must serialize as arrays and objects")
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
	for index, expected := range []map[string]any{
		{"report_time": float64(1791784485000), "updated_at": float64(1791784490123)},
		{"report_time": nil, "updated_at": nil},
	} {
		item := mustMarshalDashboardControllerShape(t, instances.InstanceList[index])
		for field, want := range expected {
			got, exists := item[field]
			if !exists || got != want {
				t.Fatalf("instance %d must preserve %s as a millisecond number or explicit null: got=%#v exists=%v want=%#v", index, field, got, exists, want)
			}
		}
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
	for _, item := range []any{node.NetworkLinks[0], instances.InstanceList[0], streams.StreamList[0], sessions.ProtocolList[0].ActiveSessionList[0]} {
		assertDashboardControllerReportPayload(t, item)
	}
	if streams.StreamList[0].ReportSubcommand != 22 || sessions.ProtocolList[0].ActiveSessionList[0].ReportSubcommand != 32 {
		t.Fatal("controller mapping lost stream or session report subcommand")
	}
	for _, item := range []any{node.ChildNodes[0].NetworkLinks[0], instances.InstanceList[1], streams.StreamList[1], sessions.ProtocolList[0].ActiveSessionList[1]} {
		payload := mustMarshalDashboardControllerShape(t, item)
		value, exists := payload["report_payload"]
		if !exists || value != nil {
			t.Fatalf("controller must preserve explicit null for legacy records: %#v", payload)
		}
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
		NetworkLinks: []*mediasvc.DashboardNetworkLink{
			{MachineId: "inst-a", SourceIp: "192.0.2.10", DestinationIp: "192.0.2.55", Rtt: 18, Jitter: 3, PacketLoss: 0.02, StatusCode: 200, ReportTime: 1791784485000, ReportPayload: dashboardControllerReportPayload()},
		},
		InstanceResources: &mediasvc.DashboardNodeResources{InstanceCount: 2, CpuAllocated: 8, CpuLoad: 2.5},
		ChildNodes: []*mediasvc.DashboardNodeOverviewItem{
			{NodeId: "node-b", NodeName: "Node B", ParentNodeId: "node-a", NodeLatencyMap: map[string]int{}, NetworkLinks: []*mediasvc.DashboardNetworkLink{{MachineId: "legacy"}}, ChildNodes: []*mediasvc.DashboardNodeOverviewItem{}},
			{NodeId: "node-c", NodeName: "Node C", ParentNodeId: "node-a", NodeLatencyMap: map[string]int{}, ChildNodes: []*mediasvc.DashboardNodeOverviewItem{}},
		},
	}}, nil
}

func (s *dashboardControllerService) ListDashboardInstances(context.Context, mediasvc.ListDashboardInstancesInput) (*mediasvc.ListDashboardInstancesOutput, error) {
	var (
		startTime  = int64(1780000000000)
		reportTime = int64(1791784485000)
		updatedAt  = int64(1791784490123)
	)
	return &mediasvc.ListDashboardInstancesOutput{
		NodeInfo: &mediasvc.DashboardInstanceNodeInfo{NodeId: "node-a", NodeName: "Node A", Region: "East", Status: "healthy"},
		List: []*mediasvc.DashboardInstanceItem{
			{InstanceId: "inst-a", InstanceName: "Transcoder A", Status: "running", StartTime: &startTime, ReportTime: &reportTime, UpdatedAt: &updatedAt, LiveStreams: 3, Sessions: 8, Version: "v1.0.0", ReportPayload: dashboardControllerReportPayload()},
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
				ReportPayload:         dashboardControllerReportPayload(),
				ReportSubcommand:      22,
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
					{SessionId: "session-a", DeviceId: "device-a", ClientId: "client-a", TenantId: "tenant-a", ProtocolType: "HLS", StartTime: &startTime, LinkHops: []*mediasvc.DashboardLinkHopItem{{HopIndex: 1, NodeId: "node-a", LatencyMs: 12}}, ReportPayload: dashboardControllerReportPayload(), ReportSubcommand: 32},
					{SessionId: "session-b", DeviceId: "device-a", ClientId: "client-b", TenantId: "tenant-a", ProtocolType: "HLS", StartTime: &startTime, LinkHops: []*mediasvc.DashboardLinkHopItem{{HopIndex: 1, NodeId: "node-b", LatencyMs: 20}}},
				},
			},
			{ProtocolType: "RTMP", Status: "active", SessionCount: 1, Sessions: []*mediasvc.DashboardSessionItem{{SessionId: "session-c", ProtocolType: "RTMP", StartTime: &startTime}}},
		},
	}, nil
}

// dashboardControllerReportPayload creates independent raw values beyond normalized DTO fields.
func dashboardControllerReportPayload() map[string]any {
	return map[string]any{
		"timestamp":   "9007199254740993",
		"packet_loss": json.Number("0.1234567890123456789"),
		"status_code": json.Number("0"),
		"extra":       map[string]any{"opaque": "before\x00after", "empty": ""},
		"link_hops": []any{map[string]any{
			"hop_id": "hop-a", "hop_name": "完整跳点", "hop_type": "gateway",
			"node_id": "4", "node_name": "节点四", "latency": json.Number("12"),
		}},
	}
}

// assertDashboardControllerReportPayload checks the serialized API boundary without float64 rounding.
func assertDashboardControllerReportPayload(t *testing.T, item any) {
	t.Helper()
	raw, err := json.Marshal(item)
	if err != nil {
		t.Fatalf("marshal controller report payload: %v", err)
	}
	var payload map[string]any
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.UseNumber()
	if err := decoder.Decode(&payload); err != nil {
		t.Fatalf("decode controller report payload: %v", err)
	}
	if !reflect.DeepEqual(payload["report_payload"], dashboardControllerReportPayload()) {
		t.Fatalf("controller changed raw report content or JSON shape: %#v", payload["report_payload"])
	}
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
