// This file verifies the tenant topology controller projection.

package media

import (
	"context"
	"testing"

	"lina-plugin-media/backend/api/media/v1"
	mediasvc "lina-plugin-media/backend/internal/service/media"
)

type dashboardTenantTopologyControllerService struct {
	dashboardControllerService
	tenantID string
}

func (s *dashboardTenantTopologyControllerService) GetDashboardTenantTopology(
	_ context.Context,
	in mediasvc.DashboardTenantTopologyInput,
) (*mediasvc.DashboardTenantTopologyOutput, error) {
	s.tenantID = in.TenantId
	return &mediasvc.DashboardTenantTopologyOutput{
		TenantId:           in.TenantId,
		ConcurrentSessions: 4,
		LiveStreamCount:    3,
		ProtocolCount:      2,
		Protocols: []*mediasvc.DashboardTenantTopologyProtocol{
			{ProtocolType: "HLS", LiveStreamCount: 2},
			{ProtocolType: "RTMP", LiveStreamCount: 1},
		},
		NodeCount:   1,
		GeneratedAt: 1780000000000,
		Nodes: []*mediasvc.DashboardTenantTopologyNode{
			{
				NodeId:             "node-a",
				NodeName:           "Node A",
				ConcurrentSessions: 4,
				LiveStreamCount:    3,
				ProtocolCount:      2,
				Protocols: []*mediasvc.DashboardTenantTopologyProtocol{
					{ProtocolType: "HLS", LiveStreamCount: 2},
					{ProtocolType: "RTMP", LiveStreamCount: 1},
				},
			},
		},
	}, nil
}

// TestGetDashboardTenantTopologyProjectsServiceOutput verifies request and response mapping.
func TestGetDashboardTenantTopologyProjectsServiceOutput(t *testing.T) {
	ctx := context.Background()
	service := &dashboardTenantTopologyControllerService{}
	controller, err := NewV1(service)
	if err != nil {
		t.Fatalf("new media controller: %v", err)
	}

	response, err := controller.GetDashboardTenantTopology(ctx, &v1.GetDashboardTenantTopologyReq{TenantId: "tenant-a"})
	if err != nil {
		t.Fatalf("get dashboard tenant topology: %v", err)
	}
	if service.tenantID != "tenant-a" {
		t.Fatalf("expected controller to pass tenant-a, got %q", service.tenantID)
	}
	payload := mustMarshalDashboardControllerShape(t, response)
	assertDashboardControllerExactKeys(t, payload,
		"tenant_id", "concurrent_sessions", "live_stream_count", "protocol_count",
		"protocols", "node_count", "node_detail_limited", "generated_at", "nodes",
	)
	if response.ConcurrentSessions != 4 || response.LiveStreamCount != 3 ||
		len(response.Nodes) != 1 || response.Nodes[0].NodeName != "Node A" {
		t.Fatalf("unexpected tenant topology response %#v", response)
	}
	assertDashboardControllerMissingKey(t, payload, "tenant_name")
}
