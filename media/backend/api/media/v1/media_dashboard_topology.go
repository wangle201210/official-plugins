// This file declares media dashboard topology DTOs.

package v1

import "github.com/gogf/gf/v2/frame/g"

// GetDashboardTopologyReq defines the request for querying the dashboard topology.
type GetDashboardTopologyReq struct {
	g.Meta   `path:"/media/dashboard/topology" method:"get" tags:"Media Dashboard" summary:"Get media gateway topology" dc:"Queries the media gateway monitoring topology by one required device GB ID and returns node, device, stream, protocol, tenant and user-session aggregates in one bounded response." permission:"media:management:query"`
	DeviceId string `json:"deviceId" v:"required|length:1,128#Device ID is required|Device ID must not exceed 128 characters" dc:"Device GB ID. This endpoint accepts only one device ID as the topology query entry." eg:"34020000001320000001"`
}

// GetDashboardTopologyRes defines the dashboard topology response.
type GetDashboardTopologyRes struct {
	NodeId               string                     `json:"node_id" dc:"Node ID that currently hosts the device. Empty when it has not been reported." eg:"node-01"`
	NodeName             string                     `json:"node_name" dc:"Node name that currently hosts the device. Empty when it has not been reported." eg:"Changsha Node"`
	DeviceId             string                     `json:"device_id" dc:"Device GB ID used by this query." eg:"34020000001320000001"`
	StreamCount          int                        `json:"stream_count" dc:"Number of active streams returned by this response." eg:"3"`
	SessionCount         int                        `json:"session_count" dc:"Total active session count in the current device query scope." eg:"12"`
	SessionDetailLimited bool                       `json:"session_detail_limited" dc:"Whether session details reached the server-side 10000 row limit. Aggregate counts still come from the database when true." eg:"false"`
	GeneratedAt          int64                      `json:"generated_at" dc:"Response generation time, Unix timestamp in milliseconds." eg:"1780923543000"`
	Devices              []*DashboardTopologyDevice `json:"devices" dc:"Device topology list." eg:"[]"`
}

// DashboardTopologyDevice defines one device subtree.
type DashboardTopologyDevice struct {
	DeviceId     string                     `json:"device_id" dc:"Device GB ID." eg:"34020000001320000001"`
	StreamCount  int                        `json:"stream_count" dc:"Number of active streams returned for this device." eg:"1"`
	SessionCount int                        `json:"session_count" dc:"Current active session count for this device." eg:"12"`
	Streams      []*DashboardTopologyStream `json:"streams" dc:"Stream topology list under this device." eg:"[]"`
}

// DashboardTopologyStream defines one stream subtree.
type DashboardTopologyStream struct {
	StreamId     string                         `json:"stream_id" dc:"Unique stream ID." eg:"stream12345"`
	StreamName   string                         `json:"stream_name" dc:"Stream display name." eg:"Camera A Stream"`
	Status       string                         `json:"status" dc:"Stream status, such as playing, paused, or error." eg:"playing"`
	BasePlatform *DashboardTopologyBasePlatform `json:"base_platform" dc:"Upstream base-platform node data." eg:"{}"`
	Gateway      *DashboardTopologyGateway      `json:"gateway" dc:"Media gateway node data." eg:"{}"`
	Protocols    []*DashboardTopologyProtocol   `json:"protocols" dc:"Protocol topology node list." eg:"[]"`
}

// DashboardTopologyBasePlatform defines the upstream platform node.
type DashboardTopologyBasePlatform struct {
	Bitrate        int    `json:"bitrate" dc:"Source bitrate in Kbps." eg:"4000"`
	Resolution     string `json:"resolution" dc:"Source resolution." eg:"1920x1080"`
	StreamProtocol string `json:"stream_protocol" dc:"Source stream protocol." eg:"RTSP"`
	DeviceCode     string `json:"device_code" dc:"Device GB code." eg:"34020000001320000001"`
	SourceUrl      string `json:"source_url" dc:"Base-platform stream URL." eg:"rtsp://example/live/stream-a"`
}

// DashboardTopologyGateway defines the video gateway node.
type DashboardTopologyGateway struct {
	StreamProtocol string  `json:"stream_protocol" dc:"Stream protocol reported by the gateway." eg:"HLS"`
	Resolution     string  `json:"resolution" dc:"Gateway output resolution." eg:"1920x1080"`
	Bitrate        int     `json:"bitrate" dc:"Gateway bitrate in Kbps." eg:"4000"`
	Fps            float64 `json:"fps" dc:"Frame rate." eg:"25"`
	AvgDelay       int     `json:"avg_delay" dc:"Average delay in milliseconds." eg:"220"`
	NodeId         string  `json:"node_id" dc:"Hosting node ID." eg:"node-01"`
	NodeName       string  `json:"node_name" dc:"Hosting node name." eg:"Changsha Node"`
	InstanceId     string  `json:"instance_id" dc:"Hosting instance ID." eg:"inst-001"`
	InstanceName   string  `json:"instance_name" dc:"Hosting instance name." eg:"media-gateway-1"`
}

// DashboardTopologyProtocol defines one protocol node under a gateway stream.
type DashboardTopologyProtocol struct {
	ProtocolType string                     `json:"protocol_type" dc:"Streaming or playback protocol type." eg:"HLS"`
	ReuseCount   int                        `json:"reuse_count" dc:"Current protocol reuse count, equal to current active sessions." eg:"6"`
	Tenants      []*DashboardTopologyTenant `json:"tenants" dc:"Tenant concurrency nodes under this protocol." eg:"[]"`
}

// DashboardTopologyTenant defines one tenant node under a protocol.
type DashboardTopologyTenant struct {
	TenantId           string                   `json:"tenant_id" dc:"Tenant ID." eg:"tenant-a"`
	ConcurrentSessions int                      `json:"concurrent_sessions" dc:"Current concurrent session count for this tenant under this protocol." eg:"12"`
	Users              []*DashboardTopologyUser `json:"users" dc:"User session nodes under this tenant." eg:"[]"`
}

// DashboardTopologyUser defines one active user session node.
type DashboardTopologyUser struct {
	SessionId    string `json:"session_id" dc:"Session ID." eg:"sess-001"`
	UserName     string `json:"user_name" dc:"User display name." eg:"viewer-a"`
	ClientId     string `json:"client_id" dc:"Client ID." eg:"client-a"`
	ClientIp     string `json:"client_ip" dc:"Client IP address." eg:"192.0.2.10"`
	ClientType   int    `json:"client_type" dc:"Client type enum: 1-mobile, 2-pc, 0-unknown." eg:"2"`
	ProtocolType string `json:"protocol_type" dc:"Playback protocol type." eg:"HLS"`
	TenantId     string `json:"tenant_id" dc:"Tenant ID." eg:"tenant-a"`
	NodeId       string `json:"node_id" dc:"Hosting node ID." eg:"node-01"`
	InstanceId   string `json:"instance_id" dc:"Hosting instance ID." eg:"inst-001"`
	StartTime    *int64 `json:"start_time" dc:"Session start time, Unix timestamp in milliseconds." eg:"1716191400000"`
	PlayDuration int    `json:"play_duration" dc:"Elapsed playback duration in seconds." eg:"1860"`
}
