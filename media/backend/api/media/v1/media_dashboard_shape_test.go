// This file verifies dashboard response DTOs keep the frontend data shape.

package v1

import (
	"bytes"
	"encoding/json"
	"reflect"
	"strings"
	"testing"
)

// TestDashboardResponseJSONShapeMatchesFrontendData verifies dashboard APIs do not add wrapper fields.
func TestDashboardResponseJSONShapeMatchesFrontendData(t *testing.T) {
	t.Run("node overview is direct node object", func(t *testing.T) {
		payload := mustMarshalDashboardShape(t, &GetDashboardNodeOverviewRes{
			NodeId:     "node-a",
			ChildNodes: []*DashboardNodeOverviewItem{},
		})
		assertDashboardHasKey(t, payload, "node_id")
		assertDashboardHasKey(t, payload, "child_nodes")
		assertDashboardMissingKey(t, payload, "nodes")
		assertDashboardMissingKey(t, payload, "total")
	})

	t.Run("instances match document data object", func(t *testing.T) {
		payload := mustMarshalDashboardShape(t, &ListDashboardInstancesRes{
			NodeInfo:     &DashboardInstanceNodeInfo{NodeId: "node-a"},
			InstanceList: []*DashboardInstanceItem{},
		})
		assertDashboardExactKeys(t, payload, "node_info", "instance_list")
		if payload["node_info"] == nil {
			t.Fatalf("expected node_info object, got nil in payload %#v", payload)
		}
	})

	t.Run("streams match document data object", func(t *testing.T) {
		payload := mustMarshalDashboardShape(t, &ListDashboardStreamsRes{
			SourceType: "node",
			SourceId:   "node-a",
			StreamList: []*DashboardStreamItem{},
		})
		assertDashboardExactKeys(t, payload, "source_type", "source_id", "stream_list")
	})

	t.Run("sessions match document data object", func(t *testing.T) {
		payload := mustMarshalDashboardShape(t, &ListDashboardSessionsRes{
			StreamInfo:   &DashboardSessionStreamInfo{StreamId: "stream-a"},
			ProtocolList: []*DashboardSessionProtocol{},
		})
		assertDashboardExactKeys(t, payload, "stream_info", "protocol_list")
	})

	t.Run("topology match gateway monitor data object", func(t *testing.T) {
		payload := mustMarshalDashboardShape(t, &GetDashboardTopologyRes{
			NodeId:   "node-a",
			DeviceId: "device-a",
			Devices: []*DashboardTopologyDevice{
				{
					DeviceId: "device-a",
					Streams: []*DashboardTopologyStream{
						{
							StreamId:     "stream-a",
							BasePlatform: &DashboardTopologyBasePlatform{},
							Gateway:      &DashboardTopologyGateway{},
							Protocols:    []*DashboardTopologyProtocol{},
						},
					},
				},
			},
		})
		assertDashboardExactKeys(t, payload,
			"node_id", "node_name", "device_id", "stream_count", "session_count",
			"session_detail_limited", "generated_at", "devices",
		)
		assertDashboardMissingKey(t, payload, "total")
	})

	t.Run("topology omits fields without report projections", func(t *testing.T) {
		assertDashboardMissingKey(t, mustMarshalDashboardShape(t, &DashboardTopologyDevice{}), "device_name")
		assertDashboardMissingKey(t, mustMarshalDashboardShape(t, &DashboardTopologyBasePlatform{}), "device_name")
		assertDashboardMissingKey(t, mustMarshalDashboardShape(t, &DashboardTopologyGateway{}), "video_code")

		user := mustMarshalDashboardShape(t, &DashboardTopologyUser{})
		assertDashboardMissingKey(t, user, "server_ip")
		assertDashboardMissingKey(t, user, "browser")
		assertDashboardMissingKey(t, user, "operating_system")
	})
}

// TestDashboardListRequestsDoNotExposePagination verifies fixed-shape list APIs are unpaged.
func TestDashboardListRequestsDoNotExposePagination(t *testing.T) {
	assertDashboardRequestMissingJSONField(t, ListDashboardInstancesReq{}, "pageNum")
	assertDashboardRequestMissingJSONField(t, ListDashboardInstancesReq{}, "pageSize")
	assertDashboardRequestMissingJSONField(t, ListDashboardStreamsReq{}, "pageNum")
	assertDashboardRequestMissingJSONField(t, ListDashboardStreamsReq{}, "pageSize")
	assertDashboardRequestMissingJSONField(t, ListDashboardSessionsReq{}, "pageNum")
	assertDashboardRequestMissingJSONField(t, ListDashboardSessionsReq{}, "pageSize")
	assertDashboardRequestMissingJSONField(t, GetDashboardTopologyReq{}, "pageNum")
	assertDashboardRequestMissingJSONField(t, GetDashboardTopologyReq{}, "pageSize")
	assertDashboardRequestMissingJSONField(t, GetDashboardTopologyReq{}, "nodeId")
	assertDashboardRequestMissingJSONField(t, GetDashboardTopologyReq{}, "deviceIds")
	assertDashboardRequestMissingJSONField(t, GetDashboardTopologyReq{}, "tenantId")
	assertDashboardRequestMissingJSONField(t, GetDashboardTopologyReq{}, "protocolType")
	assertDashboardRequestMissingJSONField(t, GetDashboardTopologyReq{}, "status")
}

// TestDashboardNetworkMetricContract preserves raw zero values and optional reachability.
func TestDashboardNetworkMetricContract(t *testing.T) {
	unknown := mustMarshalDashboardShape(t, DashboardNetworkLink{ReportTime: 1791784485000})
	assertDashboardExactKeys(t, unknown, "machine_id", "source_ip", "destination_ip", "rtt", "jitter", "packet_loss", "status_code", "reachable", "report_time", "report_payload")
	if unknown["reachable"] != nil || unknown["packet_loss"] != float64(0) || unknown["jitter"] != float64(0) || unknown["report_time"] != float64(1791784485000) {
		t.Fatalf("zero metrics and unknown reachability must be explicit: %#v", unknown)
	}
	failed := false
	payload := mustMarshalDashboardShape(t, DashboardNetworkLink{Reachable: &failed, Rtt: 5000, Jitter: 5001, PacketLoss: 1})
	if payload["reachable"] != false || payload["rtt"] != float64(5000) || payload["jitter"] != float64(5001) || payload["packet_loss"] != float64(1) {
		t.Fatalf("failure sentinels must be returned without scoring: %#v", payload)
	}
	meta, ok := reflect.TypeOf(GetDashboardNodeOverviewReq{}).FieldByName("Meta")
	if !ok || meta.Tag.Get("permission") != "media:management:query" {
		t.Fatal("network metrics must retain the existing protected overview permission")
	}
}

// TestDashboardReportPayloadJSONContract verifies complete reports stay JSON objects
// with exact raw values, while old records retain an explicit null field.
func TestDashboardReportPayloadJSONContract(t *testing.T) {
	report := map[string]any{
		"timestamp":         "9007199254740993",
		"cpu_usage":         json.Number("12.34567890123456789"),
		"status_code":       json.Number("0"),
		"watermark_enabled": false,
		"extra":             map[string]any{"unknown": "first\x00second", "empty": ""},
		"link_hops": []any{map[string]any{
			"hop_id": "hop-1", "hop_name": "完整跳点", "hop_type": "gateway",
			"node_id": "4", "node_name": "节点四", "latency": json.Number("12"),
		}},
	}
	for _, test := range []struct {
		name       string
		current    any
		legacy     any
		subcommand string
	}{
		{name: "machine", current: DashboardInstanceItem{ReportPayload: report}, legacy: DashboardInstanceItem{}},
		{name: "network", current: DashboardNetworkLink{ReportPayload: report}, legacy: DashboardNetworkLink{}},
		{name: "stream", current: DashboardStreamItem{ReportPayload: report, ReportSubcommand: 22}, legacy: DashboardStreamItem{}, subcommand: "22"},
		{name: "session", current: DashboardSessionItem{ReportPayload: report, ReportSubcommand: 32}, legacy: DashboardSessionItem{}, subcommand: "32"},
	} {
		t.Run(test.name, func(t *testing.T) {
			raw, err := json.Marshal(test.current)
			if err != nil {
				t.Fatalf("marshal populated report: %v", err)
			}
			var payload map[string]any
			decoder := json.NewDecoder(bytes.NewReader(raw))
			decoder.UseNumber()
			if err := decoder.Decode(&payload); err != nil {
				t.Fatalf("decode populated report: %v", err)
			}
			got, ok := payload["report_payload"].(map[string]any)
			if !ok || !reflect.DeepEqual(got, report) {
				t.Fatalf("report_payload must be an intact JSON object: got=%#v want=%#v", payload["report_payload"], report)
			}
			if test.subcommand != "" && payload["report_subcommand"] != json.Number(test.subcommand) {
				t.Fatalf("report subcommand changed: %#v", payload["report_subcommand"])
			}
			legacy := mustMarshalDashboardShape(t, test.legacy)
			value, exists := legacy["report_payload"]
			if !exists || value != nil {
				t.Fatalf("legacy report_payload must be present as null: %#v", legacy)
			}
			if test.subcommand != "" && legacy["report_subcommand"] != float64(0) {
				t.Fatalf("legacy subcommand must remain zero: %#v", legacy)
			}
		})
	}
}

// TestDashboardTopologyRequestContract verifies the topology entry accepts one required device ID.
func TestDashboardTopologyRequestContract(t *testing.T) {
	payload := mustMarshalDashboardShape(t, GetDashboardTopologyReq{DeviceId: "device-a"})
	assertDashboardExactKeys(t, payload, "deviceId")

	field, ok := reflect.TypeOf(GetDashboardTopologyReq{}).FieldByName("DeviceId")
	if !ok {
		t.Fatal("expected topology request device ID field")
	}
	if validation := field.Tag.Get("v"); !strings.Contains(validation, "required") {
		t.Fatalf("expected topology device ID required validation, got %q", validation)
	}
}

func mustMarshalDashboardShape(t *testing.T, value any) map[string]any {
	t.Helper()

	raw, err := json.Marshal(value)
	if err != nil {
		t.Fatalf("marshal dashboard response: %v", err)
	}
	payload := make(map[string]any)
	if err := json.Unmarshal(raw, &payload); err != nil {
		t.Fatalf("unmarshal dashboard response: %v", err)
	}
	return payload
}

func assertDashboardHasKey(t *testing.T, payload map[string]any, key string) {
	t.Helper()

	if _, ok := payload[key]; !ok {
		t.Fatalf("expected key %q in payload %#v", key, payload)
	}
}

func assertDashboardMissingKey(t *testing.T, payload map[string]any, key string) {
	t.Helper()

	if _, ok := payload[key]; ok {
		t.Fatalf("did not expect key %q in payload %#v", key, payload)
	}
}

func assertDashboardExactKeys(t *testing.T, payload map[string]any, keys ...string) {
	t.Helper()

	if len(payload) != len(keys) {
		t.Fatalf("expected keys %v, got payload %#v", keys, payload)
	}
	for _, key := range keys {
		assertDashboardHasKey(t, payload, key)
	}
}

func assertDashboardRequestMissingJSONField(t *testing.T, value any, field string) {
	t.Helper()

	raw, err := json.Marshal(value)
	if err != nil {
		t.Fatalf("marshal dashboard request: %v", err)
	}
	payload := make(map[string]any)
	if err := json.Unmarshal(raw, &payload); err != nil {
		t.Fatalf("unmarshal dashboard request: %v", err)
	}
	assertDashboardMissingKey(t, payload, field)
}
