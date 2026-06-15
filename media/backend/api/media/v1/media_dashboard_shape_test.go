// This file verifies dashboard response DTOs keep the frontend data shape.

package v1

import (
	"encoding/json"
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
}

// TestDashboardListRequestsDoNotExposePagination verifies fixed-shape list APIs are unpaged.
func TestDashboardListRequestsDoNotExposePagination(t *testing.T) {
	assertDashboardRequestMissingJSONField(t, ListDashboardInstancesReq{}, "pageNum")
	assertDashboardRequestMissingJSONField(t, ListDashboardInstancesReq{}, "pageSize")
	assertDashboardRequestMissingJSONField(t, ListDashboardStreamsReq{}, "pageNum")
	assertDashboardRequestMissingJSONField(t, ListDashboardStreamsReq{}, "pageSize")
	assertDashboardRequestMissingJSONField(t, ListDashboardSessionsReq{}, "pageNum")
	assertDashboardRequestMissingJSONField(t, ListDashboardSessionsReq{}, "pageSize")
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
