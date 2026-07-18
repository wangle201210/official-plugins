// This file declares tenant media-node topology DTOs.

package v1

import "github.com/gogf/gf/v2/frame/g"

// GetDashboardTenantTopologyReq defines the request for querying one tenant topology.
type GetDashboardTenantTopologyReq struct {
	g.Meta   `path:"/media/dashboard/tenant-topology" method:"get" tags:"Media Dashboard" summary:"Get tenant media node topology" dc:"Queries active stream and session aggregates for one required tenant ID and returns tenant totals plus bounded media-node statistics. Administrative-region filtering is not supported." permission:"media:management:query"`
	TenantId string `json:"tenantId" v:"required|length:1,64#Tenant ID is required|Tenant ID must not exceed 64 characters" dc:"Tenant ID used as the only topology query entry." eg:"tenant-a"`
}

// GetDashboardTenantTopologyRes defines the tenant topology response.
type GetDashboardTenantTopologyRes struct {
	TenantId           string                             `json:"tenant_id" dc:"Tenant ID used by this query." eg:"tenant-a"`
	ConcurrentSessions int                                `json:"concurrent_sessions" dc:"Total number of active sessions reported for this tenant." eg:"20"`
	LiveStreamCount    int                                `json:"live_stream_count" dc:"Total number of active streams reported for this tenant." eg:"8"`
	ProtocolCount      int                                `json:"protocol_count" dc:"Number of non-empty stream protocols in the tenant aggregate." eg:"2"`
	Protocols          []*DashboardTenantTopologyProtocol `json:"protocols" dc:"Tenant-level active stream counts grouped by protocol." eg:"[]"`
	NodeCount          int                                `json:"node_count" dc:"Number of media nodes returned in this bounded response." eg:"14"`
	NodeDetailLimited  bool                               `json:"node_detail_limited" dc:"Whether node aggregate details reached the server-side 10000 row or node limit. Tenant totals remain accurate when true." eg:"false"`
	GeneratedAt        int64                              `json:"generated_at" dc:"Response generation time, Unix timestamp in milliseconds." eg:"1780923543000"`
	Nodes              []*DashboardTenantTopologyNode     `json:"nodes" dc:"Bounded media-node statistics for this tenant." eg:"[]"`
}

// DashboardTenantTopologyProtocol defines active stream statistics for one protocol.
type DashboardTenantTopologyProtocol struct {
	ProtocolType    string `json:"protocol_type" dc:"Reported stream protocol." eg:"RTSP"`
	LiveStreamCount int    `json:"live_stream_count" dc:"Number of active streams reported with this protocol." eg:"5"`
}

// DashboardTenantTopologyNode defines one media-node aggregate under the tenant.
type DashboardTenantTopologyNode struct {
	NodeId             string                             `json:"node_id" dc:"Reported media node ID." eg:"node-01"`
	NodeName           string                             `json:"node_name" dc:"Latest reported media node name. Empty when no node projection exists." eg:"Changsha Node"`
	ConcurrentSessions int                                `json:"concurrent_sessions" dc:"Number of active tenant sessions reported on this node." eg:"10"`
	LiveStreamCount    int                                `json:"live_stream_count" dc:"Number of active tenant streams reported on this node." eg:"5"`
	ProtocolCount      int                                `json:"protocol_count" dc:"Number of non-empty stream protocols in this node aggregate." eg:"2"`
	Protocols          []*DashboardTenantTopologyProtocol `json:"protocols" dc:"Node-level active stream counts grouped by protocol." eg:"[]"`
}
