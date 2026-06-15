// This file declares media dashboard stream-list DTOs.

package v1

import "github.com/gogf/gf/v2/frame/g"

// ListDashboardStreamsReq defines the request for querying dashboard streams.
type ListDashboardStreamsReq struct {
	g.Meta     `path:"/media/dashboard/streams" method:"get" tags:"媒体数据看板" summary:"查询流信息列表" dc:"查询媒体数据看板流信息列表，支持按来源、租户、节点、实例、状态和关键词筛选；最多返回10000条。" permission:"media:management:query"`
	SourceType string `json:"sourceType" dc:"数据来源：node节点或instance实例" eg:"node"`
	SourceId   string `json:"sourceId" dc:"来源ID，节点ID或实例ID" eg:"node-01"`
	TenantId   string `json:"tenantId" dc:"按租户ID筛选" eg:"tenant-a"`
	NodeId     string `json:"nodeId" dc:"按节点ID筛选" eg:"node-01"`
	InstanceId string `json:"instanceId" dc:"按实例ID筛选" eg:"inst-001"`
	Status     string `json:"status" dc:"按流状态筛选，例如playing、paused、error" eg:"playing"`
	Keyword    string `json:"keyword" dc:"按流ID、流名称、源地址、节点名称或实例名称模糊筛选" eg:"camera"`
}

// ListDashboardStreamsRes defines the dashboard stream-list response.
type ListDashboardStreamsRes struct {
	SourceType string                 `json:"source_type" dc:"数据来源：node节点或instance实例" eg:"node"`
	SourceId   string                 `json:"source_id" dc:"来源ID，节点ID或实例ID" eg:"node-01"`
	StreamList []*DashboardStreamItem `json:"stream_list" dc:"流列表" eg:"[]"`
}

// DashboardStreamItem defines one dashboard stream row.
type DashboardStreamItem struct {
	SourceUrl             string                   `json:"source_url" dc:"基础平台流地址" eg:"xit1Ic2K9MnGd.flv"`
	StreamId              string                   `json:"stream_id" dc:"流唯一标识" eg:"stream12345"`
	StreamName            string                   `json:"stream_name" dc:"流名称" eg:"XX摄像头流"`
	Resolution            string                   `json:"resolution" dc:"原始分辨率" eg:"1920x1080"`
	Fps                   float64                  `json:"fps" dc:"原始帧率" eg:"25"`
	Bitrate               int                      `json:"bitrate" dc:"原始码率，单位kbps" eg:"4000"`
	PacketLoss            float64                  `json:"packet_loss" dc:"源流丢包率" eg:"0.12"`
	Status                string                   `json:"status" dc:"流状态，例如playing、paused、error" eg:"playing"`
	StartTime             *int64                   `json:"start_time" dc:"流开始时间，Unix timestamp in milliseconds" eg:"1716187200000"`
	Duration              int                      `json:"duration" dc:"已运行时长，单位秒" eg:"7200"`
	AvgDelay              int                      `json:"avg_delay" dc:"源流延迟，单位毫秒" eg:"220"`
	ProtocolCount         int                      `json:"protocol_count" dc:"支持的协议数" eg:"2"`
	TotalSessionsLifetime int64                    `json:"total_sessions_lifetime" dc:"流存活期间累计会话数" eg:"1250"`
	CurrentActiveSessions int                      `json:"current_active_sessions" dc:"当前仍在线的会话数" eg:"85"`
	WatermarkEnabled      bool                     `json:"watermark_enabled" dc:"是否启用水印" eg:"true"`
	ProtocolSummary       []*DashboardProtocolItem `json:"protocol_summary" dc:"按协议统计" eg:"[]"`
}

// DashboardProtocolItem defines one dashboard protocol summary row.
type DashboardProtocolItem struct {
	ProtocolType    string `json:"protocol_type" dc:"协议类型，例如RTMP、HLS、FLV、DASH" eg:"RTMP"`
	TotalSessions   int    `json:"total_sessions" dc:"协议历史累计会话数" eg:"3000"`
	CurrentSessions int    `json:"current_sessions" dc:"协议当前会话数" eg:"30"`
}
