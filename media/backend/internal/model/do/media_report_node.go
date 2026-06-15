// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"
)

// MediaReportNode is the golang structure of table media_report_node for DAO operations like Where/Data.
type MediaReportNode struct {
	g.Meta          `orm:"table:media_report_node, do:true"`
	NodeId          any         // 上报节点业务标识，不依赖配置表外键
	NodeName        any         // 节点展示名称，按上报时间点冗余
	Region          any         // 节点所属区域或机房名称
	ParentNodeId    any         // 父节点业务标识，根节点使用0
	Status          any         // 节点运行状态，保存上报原始枚举值
	TotalNodes      any         // 当前节点视角下的节点总数
	AliveNodes      any         // 当前节点视角下的在线节点数
	CpuAllocated    any         // 已分配CPU容量，单位由上报端统一
	CpuLoad         any         // CPU负载或利用率，单位由上报端统一
	MemoryAllocated any         // 已分配内存容量，单位GB
	MemoryUsed      any         // 已使用内存容量，单位GB
	DiskIoRead      any         // 磁盘读取速率，单位MB/s
	DiskIoWrite     any         // 磁盘写入速率，单位MB/s
	NetworkIn       any         // 网络入站速率，单位Mbps
	NetworkOut      any         // 网络出站速率，单位Mbps
	LiveStreams     any         // 当前直播流数量
	Sessions        any         // 当前会话数量
	AvgDelay        any         // 节点平均延迟，单位毫秒
	LastHeartbeat   *gtime.Time // 节点最后心跳时间
	NodeLatencyMap  any         // 节点到其他节点的延迟矩阵JSON，结构来自接口node_latency_map
	ReportTime      any         // 上报端采样时间
	UpdatedAt       *gtime.Time // 记录更新时间
}
