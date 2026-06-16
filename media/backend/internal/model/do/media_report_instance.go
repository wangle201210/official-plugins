// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"
)

// MediaReportInstance is the golang structure of table media_report_instance for DAO operations like Where/Data.
type MediaReportInstance struct {
	g.Meta          `orm:"table:media_report_instance, do:true"`
	InstanceId      any         // 实例业务标识，不依赖实例配置外键
	InstanceName    any         // 实例展示名称，按上报时间点冗余
	NodeId          any         // 实例所属节点业务标识
	NodeName        any         // 实例所属节点展示名称，按上报时间点冗余
	Region          any         // 实例所属区域或机房名称
	NodeStatus      any         // 实例所属节点状态，用于返回node_info.status
	Status          any         // 实例运行状态，保存上报原始枚举值
	CpuAllocated    any         // 实例已分配CPU容量，单位由上报端统一
	CpuLoad         any         // 实例CPU负载或利用率，单位由上报端统一
	MemoryAllocated any         // 实例已分配内存容量，单位MB
	MemoryUsed      any         // 实例已使用内存容量，单位MB
	DiskIoRead      any         // 实例磁盘读取速率，单位KB/S
	DiskIoWrite     any         // 实例磁盘写入速率，单位KB/S
	NetworkIn       any         // 实例网络入站速率，单位KB/S
	NetworkOut      any         // 实例网络出站速率，单位KB/S
	LiveStreams     any         // 实例当前承载直播流数量
	Sessions        any         // 实例当前会话数量
	StartTime       *gtime.Time // 实例启动时间
	Version         any         // 实例版本号
	ReportTime      any         // 上报端采样时间
	UpdatedAt       *gtime.Time // 记录更新时间
}
