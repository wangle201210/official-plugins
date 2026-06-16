// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

import (
	"github.com/gogf/gf/v2/os/gtime"
)

// MediaReportStream is the golang structure for table media_report_stream.
type MediaReportStream struct {
	StreamId              string      `json:"streamId"              orm:"stream_id"               description:"流业务标识，不依赖流配置外键"`
	SourceType            string      `json:"sourceType"            orm:"source_type"             description:"来源类型，例如节点或实例"`
	TenantId              string      `json:"tenantId"              orm:"tenant_id"               description:"租户标识，用于数据权限过滤和租户统计"`
	NodeId                string      `json:"nodeId"                orm:"node_id"                 description:"流所属节点业务标识"`
	NodeName              string      `json:"nodeName"              orm:"node_name"               description:"流所属节点展示名称，按上报时间点冗余"`
	InstanceId            string      `json:"instanceId"            orm:"instance_id"             description:"流所属实例业务标识"`
	InstanceName          string      `json:"instanceName"          orm:"instance_name"           description:"流所属实例展示名称，按上报时间点冗余"`
	SourceUrl             string      `json:"sourceUrl"             orm:"source_url"              description:"流源地址"`
	StreamName            string      `json:"streamName"            orm:"stream_name"             description:"流展示名称"`
	Resolution            string      `json:"resolution"            orm:"resolution"              description:"当前分辨率"`
	Fps                   float64     `json:"fps"                   orm:"fps"                     description:"当前帧率"`
	Bitrate               int         `json:"bitrate"               orm:"bitrate"                 description:"当前码率，单位Kbps"`
	PacketLoss            float64     `json:"packetLoss"            orm:"packet_loss"             description:"当前丢包率"`
	Status                string      `json:"status"                orm:"status"                  description:"流运行状态，保存上报原始枚举值"`
	StartTime             *gtime.Time `json:"startTime"             orm:"start_time"              description:"流开始时间"`
	Duration              int         `json:"duration"              orm:"duration"                description:"上报端流持续时间，接口返回时优先通过start_time和close_time动态计算"`
	AvgDelay              int         `json:"avgDelay"              orm:"avg_delay"               description:"流平均延迟，单位毫秒"`
	ProtocolCount         int         `json:"protocolCount"         orm:"protocol_count"          description:"支持的协议数量"`
	TotalSessionsLifetime int64       `json:"totalSessionsLifetime" orm:"total_sessions_lifetime" description:"流历史累计会话数量"`
	CurrentActiveSessions int         `json:"currentActiveSessions" orm:"current_active_sessions" description:"流当前活跃会话数量"`
	WatermarkEnabled      bool        `json:"watermarkEnabled"      orm:"watermark_enabled"       description:"当前是否启用水印"`
	ProtocolSummary       string      `json:"protocolSummary"       orm:"protocol_summary"        description:"协议摘要JSON数组，结构来自接口protocol_summary"`
	ReportTime            int64       `json:"reportTime"            orm:"report_time"             description:"上报端采样时间"`
	UpdatedAt             *gtime.Time `json:"updatedAt"             orm:"updated_at"              description:"记录更新时间"`
	ProtocolType          string      `json:"protocolType"          orm:"protocol_type"           description:"源流协议类型"`
	CloseTime             *gtime.Time `json:"closeTime"             orm:"close_time"              description:"流关闭时间，未关闭时为空"`
}
