// This file declares media dashboard session-list DTOs.

package v1

import "github.com/gogf/gf/v2/frame/g"

// ListDashboardSessionsReq defines the request for querying dashboard sessions.
type ListDashboardSessionsReq struct {
	g.Meta       `path:"/media/dashboard/sessions" method:"get" tags:"媒体数据看板" summary:"查询会话列表" dc:"按必填streamId查询未关闭会话，只返回close_time IS NULL记录，不返回已关闭会话，不按心跳超时判断实际播放。可选筛选条件取交集。明细按协议升序、上报时间降序、会话ID升序读取，所有协议共享最多10000条明细，无分页和total；session_count为对应筛选范围的完整计数，可能大于明细长度。结果位于统一响应code、message、data的data内。" permission:"media:management:query"`
	StreamId     string `json:"streamId" v:"required#流ID不能为空" dc:"必填，按流ID精确查询；未传或去除首尾空白后为空时返回参数错误" eg:"stream12345"`
	TenantId     string `json:"tenantId" dc:"可选，按租户客户编码精确筛选；不传时不限制租户" eg:"tenant-a"`
	DeviceId     string `json:"deviceId" dc:"可选，按设备国标ID精确筛选；不传时不限制设备" eg:"34020000001320000001"`
	ProtocolType string `json:"protocolType" dc:"可选，按协议存储字符串精确匹配；不传时包含所有协议。采集标准值：UNIVERSAL未知、RTMP、RTSP、HLS、HTTP_FLV、WS_FLV、HTTPS_FLV、WSS_FLV、GB28181；不做枚举校验，不自动将FLV映射为HTTP_FLV" eg:"RTMP"`
	NodeId       string `json:"nodeId" dc:"可选，按会话所在节点ID精确筛选，不包含子节点；不传时不限制节点" eg:"node-01"`
	InstanceId   string `json:"instanceId" dc:"可选，按会话所在实例ID精确筛选；不传时不限制实例" eg:"inst-001"`
	Keyword      string `json:"keyword" dc:"可选，按会话ID、设备ID、客户端ID、客户端IP、用户名或协议类型模糊匹配；不传时不增加关键词条件" eg:"sess"`
}

// ListDashboardSessionsRes defines the dashboard session-list response.
type ListDashboardSessionsRes struct {
	StreamInfo   *DashboardSessionStreamInfo `json:"stream_info" dc:"流基础信息" eg:"{}"`
	ProtocolList []*DashboardSessionProtocol `json:"protocol_list" dc:"协议分组；可能保留流上报摘要中的零会话协议，不能用分组个数判断活跃会话数" eg:"[]"`
}

// DashboardSessionStreamInfo defines the stream summary returned with sessions.
type DashboardSessionStreamInfo struct {
	StreamId   string `json:"stream_id" dc:"流唯一标识" eg:"stream12345"`
	StreamName string `json:"stream_name" dc:"流名称" eg:"XX摄像头流"`
	DeviceId   string `json:"device_id" dc:"设备国标ID" eg:"34020000001320000001"`
}

// DashboardSessionProtocol defines one protocol group and its active sessions.
type DashboardSessionProtocol struct {
	ProtocolType      string                  `json:"protocol_type" dc:"协议存储字符串；采集标准值：UNIVERSAL未知、RTMP、RTSP、HLS、HTTP_FLV、WS_FLV、HTTPS_FLV、WSS_FLV、GB28181；其他存储值按原值返回" eg:"RTMP"`
	Status            string                  `json:"status" dc:"当前筛选范围的协议状态：active表示session_count大于0，inactive表示等于0；不表示流的status" eg:"active"`
	SessionCount      int                     `json:"session_count" dc:"该协议在当前筛选范围内的完整未关闭会话数，不受明细10000条上限影响，可能大于active_session_list长度" eg:"30"`
	ActiveSessionList []*DashboardSessionItem `json:"active_session_list" dc:"当前协议下未关闭会话明细；所有协议共享10000条上限，无明细时为空数组，即使计数大于0也可能因截断而为空" eg:"[]"`
}

// DashboardSessionItem defines one dashboard session row.
type DashboardSessionItem struct {
	SessionId         string                  `json:"session_id" dc:"全局唯一会话ID" eg:"sess-9f3a1c8e"`
	DeviceId          string                  `json:"device_id" dc:"设备国标ID" eg:"34020000001320000001"`
	ClientId          string                  `json:"client_id" dc:"客户端身份" eg:"user-88721"`
	ClientIp          string                  `json:"client_ip" dc:"客户端IP" eg:"1.202.33.41"`
	ClientType        int                     `json:"client_type" dc:"客户端类型枚举：1-mobile，2-pc，0-未知" eg:"1"`
	TenantId          string                  `json:"tenant_id" dc:"拉流租户ID" eg:"12145"`
	UserName          string                  `json:"user_name" dc:"拉流用户" eg:"张三"`
	ProtocolType      string                  `json:"protocol_type" dc:"播放协议存储字符串；采集标准值：UNIVERSAL未知、RTMP、RTSP、HLS、HTTP_FLV、WS_FLV、HTTPS_FLV、WSS_FLV、GB28181" eg:"RTMP"`
	StartTime         *int64                  `json:"start_time" dc:"会话开始时间，Unix timestamp in milliseconds" eg:"1716191400000"`
	PlayDuration      int                     `json:"play_duration" dc:"会话持续时长，单位秒；优先根据开始时间计算至当前时间，无法计算时回退为上报时长；不等同于连续有画面的时长" eg:"1860"`
	CurrentFps        float64                 `json:"current_fps" dc:"当前fps" eg:"25"`
	CurrentBitrate    int                     `json:"current_bitrate" dc:"当前码率，单位kbps" eg:"2500"`
	CurrentResolution string                  `json:"current_resolution" dc:"当前分辨率" eg:"854x480"`
	NodeId            string                  `json:"node_id" dc:"当前接入节点" eg:"node-01"`
	InstanceId        string                  `json:"instance_id" dc:"当前处理实例" eg:"inst-003"`
	LinkHops          []*DashboardLinkHopItem `json:"link_hops" dc:"会话实际链路" eg:"[]"`
	TotalLinkLatency  int                     `json:"total_link_latency" dc:"链路总延迟，单位毫秒" eg:"28"`
	ReportPayload     map[string]any          `json:"report_payload" dc:"完整最新SessionMetric原始JSON对象，包含全部已定义字段、完整extra和link_hops；没有原始上报时为null。遵循protobuf JSON，64位整数为字符串，时间保留上报单位" eg:"{\"session_id\":\"sess-9f3a1c8e\",\"link_hops\":[],\"extra\":{}}"`
	ReportSubcommand  int                     `json:"report_subcommand" dc:"原始会话上报子命令：0未记录或保留值，30新增，31删除，32状态，33失败，40会话查询请求，41会话查询响应；其他值保留上报原值" eg:"32"`
}

// DashboardLinkHopItem defines one session link hop.
type DashboardLinkHopItem struct {
	HopIndex  int    `json:"hop_index" dc:"链路跳点序号" eg:"1"`
	NodeId    string `json:"node_id" dc:"跳点节点ID" eg:"origin-node"`
	LatencyMs int    `json:"latency_ms" dc:"跳点延迟，单位毫秒" eg:"10"`
}
