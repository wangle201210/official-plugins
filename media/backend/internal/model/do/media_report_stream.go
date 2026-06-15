// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"
)

// MediaReportStream is the golang structure of table media_report_stream for DAO operations like Where/Data.
type MediaReportStream struct {
	g.Meta                `orm:"table:media_report_stream, do:true"`
	StreamId              any         // 流业务标识，不依赖流配置外键
	SourceType            any         // 来源类型，例如节点或实例
	SourceId              any         // 来源业务标识，用于节点或实例下钻
	TenantId              any         // 租户标识，用于数据权限过滤和租户统计
	NodeId                any         // 流所属节点业务标识
	NodeName              any         // 流所属节点展示名称，按上报时间点冗余
	InstanceId            any         // 流所属实例业务标识
	InstanceName          any         // 流所属实例展示名称，按上报时间点冗余
	SourceUrl             any         // 流源地址
	StreamName            any         // 流展示名称
	Resolution            any         // 当前分辨率
	Fps                   any         // 当前帧率
	Bitrate               any         // 当前码率，单位Kbps
	PacketLoss            any         // 当前丢包率
	Status                any         // 流运行状态，保存上报原始枚举值
	StartTime             *gtime.Time // 流开始时间
	Duration              any         // 流持续时间，单位秒
	AvgDelay              any         // 流平均延迟，单位毫秒
	ProtocolCount         any         // 支持的协议数量
	TotalSessionsLifetime any         // 流历史累计会话数量
	CurrentActiveSessions any         // 流当前活跃会话数量
	WatermarkEnabled      any         // 当前是否启用水印
	ProtocolSummary       any         // 协议摘要JSON数组，结构来自接口protocol_summary
	ReportTime            any         // 上报端采样时间
	UpdatedAt             *gtime.Time // 记录更新时间
}
