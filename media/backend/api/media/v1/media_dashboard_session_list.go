// This file declares media dashboard session-list DTOs.

package v1

import "github.com/gogf/gf/v2/frame/g"

// ListDashboardSessionsReq defines the request for querying dashboard sessions.
type ListDashboardSessionsReq struct {
	g.Meta       `path:"/media/dashboard/sessions" method:"get" tags:"媒体数据看板" summary:"查询会话列表" dc:"查询媒体数据看板会话列表，按流和可选维度组装协议分组；最多返回10000条会话明细。" permission:"media:management:query"`
	StreamId     string `json:"streamId" v:"required#流ID不能为空" dc:"按流ID查询会话" eg:"stream12345"`
	TenantId     string `json:"tenantId" dc:"按租户ID筛选" eg:"tenant-a"`
	ProtocolType string `json:"protocolType" dc:"按协议类型筛选，例如RTMP、HLS、FLV、DASH" eg:"RTMP"`
	NodeId       string `json:"nodeId" dc:"按节点ID筛选" eg:"node-01"`
	InstanceId   string `json:"instanceId" dc:"按实例ID筛选" eg:"inst-001"`
	Keyword      string `json:"keyword" dc:"按会话ID、客户端ID、客户端IP、用户名或协议类型模糊筛选" eg:"sess"`
}

// ListDashboardSessionsRes defines the dashboard session-list response.
type ListDashboardSessionsRes struct {
	StreamInfo   *DashboardSessionStreamInfo `json:"stream_info" dc:"流基础信息" eg:"{}"`
	ProtocolList []*DashboardSessionProtocol `json:"protocol_list" dc:"协议列表" eg:"[]"`
}

// DashboardSessionStreamInfo defines the stream summary returned with sessions.
type DashboardSessionStreamInfo struct {
	StreamId   string `json:"stream_id" dc:"流唯一标识" eg:"stream12345"`
	StreamName string `json:"stream_name" dc:"流名称" eg:"XX摄像头流"`
}

// DashboardSessionProtocol defines one protocol group and its active sessions.
type DashboardSessionProtocol struct {
	ProtocolType      string                  `json:"protocol_type" dc:"协议类型，例如RTMP、HLS、FLV、DASH" eg:"RTMP"`
	Status            string                  `json:"status" dc:"协议状态：active或inactive" eg:"active"`
	SessionCount      int                     `json:"session_count" dc:"该协议下当前查询范围内会话数" eg:"30"`
	ActiveSessionList []*DashboardSessionItem `json:"active_session_list" dc:"当前协议下的会话列表" eg:"[]"`
}

// DashboardSessionItem defines one dashboard session row.
type DashboardSessionItem struct {
	SessionId         string                  `json:"session_id" dc:"全局唯一会话ID" eg:"sess-9f3a1c8e"`
	ClientId          string                  `json:"client_id" dc:"客户端身份" eg:"user-88721"`
	ClientIp          string                  `json:"client_ip" dc:"客户端IP" eg:"1.202.33.41"`
	ClientType        int                     `json:"client_type" dc:"客户端类型枚举：1-mobile，2-pc，0-未知" eg:"1"`
	TenantId          string                  `json:"tenant_id" dc:"拉流租户ID" eg:"12145"`
	UserName          string                  `json:"user_name" dc:"拉流用户" eg:"张三"`
	ProtocolType      string                  `json:"protocol_type" dc:"播放协议类型" eg:"RTMP"`
	StartTime         *int64                  `json:"start_time" dc:"会话开始时间，Unix timestamp in milliseconds" eg:"1716191400000"`
	PlayDuration      int                     `json:"play_duration" dc:"已播放时长，单位秒" eg:"1860"`
	CurrentFps        float64                 `json:"current_fps" dc:"当前fps" eg:"25"`
	CurrentBitrate    int                     `json:"current_bitrate" dc:"当前码率，单位kbps" eg:"2500"`
	CurrentResolution string                  `json:"current_resolution" dc:"当前分辨率" eg:"854x480"`
	NodeId            string                  `json:"node_id" dc:"当前接入节点" eg:"node-01"`
	InstanceId        string                  `json:"instance_id" dc:"当前处理实例" eg:"inst-003"`
	LinkHops          []*DashboardLinkHopItem `json:"link_hops" dc:"会话实际链路" eg:"[]"`
	TotalLinkLatency  int                     `json:"total_link_latency" dc:"链路总延迟，单位毫秒" eg:"28"`
}

// DashboardLinkHopItem defines one session link hop.
type DashboardLinkHopItem struct {
	HopIndex  int    `json:"hop_index" dc:"链路跳点序号" eg:"1"`
	NodeId    string `json:"node_id" dc:"跳点节点ID" eg:"origin-node"`
	LatencyMs int    `json:"latency_ms" dc:"跳点延迟，单位毫秒" eg:"10"`
}
