// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"
)

// MediaReportNodeSnapshot is the golang structure of table media_report_node_snapshot for DAO operations like Where/Data.
type MediaReportNodeSnapshot struct {
	g.Meta          `orm:"table:media_report_node_snapshot, do:true"`
	NodeId          any         // 上报节点业务标识，不依赖配置表外键
	ReportTime      any         // 上报端采样时间，也是快照时间
	NodeName        any         // 节点展示名称，按快照时间点冗余
	Region          any         // 节点所属区域或机房名称
	ParentNodeId    any         // 父节点业务标识，根节点使用0
	Status          any         // 节点运行状态，保存上报原始枚举值
	TotalNodes      any         // 快照时节点视角下的节点总数
	AliveNodes      any         // 快照时节点视角下的在线节点数
	CpuAllocated    any         // 快照时已分配CPU容量，单位由上报端统一
	CpuLoad         any         // 快照时CPU负载或利用率，单位由上报端统一
	MemoryAllocated any         // 快照时已分配内存容量，单位MB
	MemoryUsed      any         // 快照时已使用内存容量，单位MB
	DiskIoRead      any         // 快照时磁盘读取速率，单位KB/S
	DiskIoWrite     any         // 快照时磁盘写入速率，单位KB/S
	NetworkIn       any         // 快照时网络入站速率，单位KB/S
	NetworkOut      any         // 快照时网络出站速率，单位KB/S
	LiveStreams     any         // 快照时直播流数量
	Sessions        any         // 快照时会话数量
	AvgDelay        any         // 快照时节点平均延迟，单位毫秒
	LastHeartbeat   *gtime.Time // 快照时节点最后心跳时间
	NodeLatencyMap  any         // 快照时节点延迟矩阵JSON，结构来自接口node_latency_map
}
