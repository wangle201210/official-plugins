// This file verifies the tenant topology request and response contract.

package v1

import (
	"reflect"
	"strings"
	"testing"
)

// TestDashboardTenantTopologyContract verifies the endpoint exposes only tenant-backed fields.
func TestDashboardTenantTopologyContract(t *testing.T) {
	request := mustMarshalDashboardShape(t, GetDashboardTenantTopologyReq{TenantId: "tenant-a"})
	assertDashboardExactKeys(t, request, "tenantId")

	field, ok := reflect.TypeOf(GetDashboardTenantTopologyReq{}).FieldByName("TenantId")
	if !ok {
		t.Fatal("expected tenant topology request tenant ID field")
	}
	validation := field.Tag.Get("v")
	if !strings.Contains(validation, "required") || !strings.Contains(validation, "length:1,64") {
		t.Fatalf("expected required tenant ID validation, got %q", validation)
	}

	response := mustMarshalDashboardShape(t, &GetDashboardTenantTopologyRes{
		TenantId:  "tenant-a",
		Protocols: []*DashboardTenantTopologyProtocol{},
		Nodes:     []*DashboardTenantTopologyNode{},
	})
	assertDashboardExactKeys(t, response,
		"tenant_id", "concurrent_sessions", "live_stream_count", "protocol_count",
		"protocols", "node_count", "node_detail_limited", "generated_at", "nodes",
	)
	assertDashboardMissingKey(t, response, "tenant_name")
	assertDashboardMissingKey(t, response, "region")
	assertDashboardMissingKey(t, response, "administrative_region")
}
