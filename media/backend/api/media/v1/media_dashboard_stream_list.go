// This file declares media dashboard stream-list DTOs.

package v1

import "github.com/gogf/gf/v2/frame/g"

// ListDashboardStreamsReq defines the request for querying dashboard streams.
type ListDashboardStreamsReq struct {
	g.Meta     `path:"/media/dashboard/streams" method:"get" tags:"媒体数据看板" summary:"查询流信息列表" dc:"查询最新流投影，默认包含未关闭和已关闭记录；可按来源、租户、设备、节点、实例、状态和关键词组合筛选。按上报时间降序、流ID升序返回，最多10000条，无分页和total。status=running仅筛选上报状态；summary还要求流未关闭；已关闭超过30天的记录由定时任务清理。结果位于统一响应code、message、data的data内。" permission:"media:management:query"`
	SourceType string `json:"sourceType" dc:"可选，按来源类型精确筛选：node节点、instance实例；不传时不限制来源类型。与sourceId组合使用；显式node还限制来源类型，不等同于单独传nodeId" eg:"node"`
	SourceId   string `json:"sourceId" dc:"可选，sourceType=instance时匹配实例ID；sourceType=node或未传时匹配节点ID。不传时不增加来源ID条件；与nodeId、instanceId等条件取交集" eg:"node-01"`
	TenantId   string `json:"tenantId" dc:"可选，按租户客户编码精确筛选；不传时不限制租户" eg:"tenant-a"`
	DeviceId   string `json:"deviceId" dc:"可选，按设备国标ID精确筛选；不传时不限制设备" eg:"34020000001320000001"`
	NodeId     string `json:"nodeId" dc:"可选，按直属节点ID精确筛选，不包含子节点；不传时不限制节点" eg:"node-01"`
	InstanceId string `json:"instanceId" dc:"可选，按实例ID精确筛选；不传时不限制实例" eg:"inst-001"`
	Status     string `json:"status" dc:"可选，去除首尾空白后按存储状态精确匹配；不传或为空时不过滤状态。采集标准值：universal未知或未指定、running运行中、inactive非活动、error异常、closed已关闭、timeout超时、cancelled已取消、failed失败。不做枚举校验，不自动映射playing或paused" eg:"running"`
	Keyword    string `json:"keyword" dc:"可选，按流ID、流名称、源地址、设备ID、节点名称或实例名称模糊匹配；不传时不增加关键词条件" eg:"camera"`
}

// ListDashboardStreamsRes defines the dashboard stream-list response.
type ListDashboardStreamsRes struct {
	SourceType string                 `json:"source_type" dc:"去除首尾空白后的请求sourceType回显；未传时为空，不从结果推断" eg:"node"`
	SourceId   string                 `json:"source_id" dc:"去除首尾空白后的请求sourceId回显；未传时为空，不从结果推断" eg:"node-01"`
	StreamList []*DashboardStreamItem `json:"stream_list" dc:"符合筛选的流列表，最多10000条；无匹配时为空数组，包含尚未清理的已关闭流" eg:"[]"`
}

// DashboardStreamItem defines one dashboard stream row.
type DashboardStreamItem struct {
	SourceUrl             string                   `json:"source_url" dc:"上报stream_path，缺省回退stream_alias；可能为相对路径或别名，不保证是绝对URL" eg:"xit1Ic2K9MnGd.flv"`
	StreamId              string                   `json:"stream_id" dc:"流唯一标识" eg:"stream12345"`
	StreamName            string                   `json:"stream_name" dc:"流展示名称，优先使用上报stream_alias，缺省使用stream_id" eg:"XX摄像头流"`
	DeviceId              string                   `json:"device_id" dc:"设备国标ID" eg:"34020000001320000001"`
	Resolution            string                   `json:"resolution" dc:"原始分辨率" eg:"1920x1080"`
	Fps                   float64                  `json:"fps" dc:"原始帧率" eg:"25"`
	Bitrate               int                      `json:"bitrate" dc:"原始码率，单位kbps" eg:"4000"`
	PacketLoss            float64                  `json:"packet_loss" dc:"源流丢包率" eg:"0.12"`
	Status                string                   `json:"status" dc:"最新存储状态；采集标准值：universal未知或未指定、running运行中、inactive非活动、error异常、closed已关闭、timeout超时、cancelled已取消、failed失败。summary要求此字段为running且流未关闭" eg:"running"`
	StartTime             *int64                   `json:"start_time" dc:"流开始时间，Unix timestamp in milliseconds" eg:"1716187200000"`
	Duration              int                      `json:"duration" dc:"流持续时长，单位秒；优先按启动至关闭时间计算，未关闭时计算至当前时间，无法计算时回退为上报时长" eg:"7200"`
	AvgDelay              int                      `json:"avg_delay" dc:"源流延迟，单位毫秒" eg:"220"`
	ProtocolCount         int                      `json:"protocol_count" dc:"合并上报摘要与当前未关闭会话后的非空协议种类数；无可用摘要时回退流上报值" eg:"2"`
	TotalSessionsLifetime int64                    `json:"total_sessions_lifetime" dc:"累计会话数，来自上报摘要或流投影；不是当前活跃会话数，不受列表上限影响" eg:"1250"`
	CurrentActiveSessions int                      `json:"current_active_sessions" dc:"优先汇总该流各协议的未关闭会话数；无协议摘要时回退流上报值。会话聚合仅按stream_id，不重复应用流列表的租户等筛选；不受会话明细10000条上限影响，不根据心跳判断实际播放" eg:"85"`
	WatermarkEnabled      bool                     `json:"watermark_enabled" dc:"是否启用水印" eg:"true"`
	ProtocolSummary       []*DashboardProtocolItem `json:"protocol_summary" dc:"协议摘要，合并上报累计值和当前未关闭会话计数；无数据时为空数组" eg:"[]"`
	ReportPayload         map[string]any           `json:"report_payload" dc:"完整最新StreamMetric原始JSON对象，包含编码、服务器类型、全部已定义字段和extra；没有原始上报时为null。遵循protobuf JSON，64位整数为字符串，时间保留上报单位" eg:"{\"stream_id\":\"stream12345\",\"video_codec\":1,\"extra\":{}}"`
	ReportSubcommand      int                      `json:"report_subcommand" dc:"原始流上报子命令：0未记录或保留值，20新增，21删除，22状态，23失败，24流查询请求，25流查询响应；其他值保留上报原值" eg:"22"`
}

// DashboardProtocolItem defines one dashboard protocol summary row.
type DashboardProtocolItem struct {
	ProtocolType    string `json:"protocol_type" dc:"协议字符串；采集标准值：UNIVERSAL未知、RTMP、RTSP、HLS、HTTP_FLV、WS_FLV、HTTPS_FLV、WSS_FLV、GB28181；其他存储值按原值返回" eg:"RTMP"`
	TotalSessions   int    `json:"total_sessions" dc:"该协议上报的累计会话数；仅从活跃会话补出的协议没有累计值时为0" eg:"3000"`
	CurrentSessions int    `json:"current_sessions" dc:"该流该协议的未关闭会话数，以会话close_time IS NULL统计" eg:"30"`
}
