// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

import (
	"github.com/gogf/gf/v2/os/gtime"
)

// MediaReportInstance is the golang structure for table media_report_instance.
type MediaReportInstance struct {
	InstanceId      string      `json:"instanceId"      orm:"instance_id"      description:"实例业务标识，不依赖实例配置外键"`
	InstanceName    string      `json:"instanceName"    orm:"instance_name"    description:"实例展示名称，按上报时间点冗余"`
	NodeId          string      `json:"nodeId"          orm:"node_id"          description:"实例所属节点业务标识"`
	NodeName        string      `json:"nodeName"        orm:"node_name"        description:"实例所属节点展示名称，按上报时间点冗余"`
	Region          string      `json:"region"          orm:"region"           description:"实例所属区域或机房名称"`
	NodeStatus      string      `json:"nodeStatus"      orm:"node_status"      description:"实例所属节点状态，用于返回node_info.status"`
	Status          string      `json:"status"          orm:"status"           description:"实例运行状态，保存上报原始枚举值"`
	CpuAllocated    float64     `json:"cpuAllocated"    orm:"cpu_allocated"    description:"实例已分配CPU容量，单位由上报端统一"`
	CpuLoad         float64     `json:"cpuLoad"         orm:"cpu_load"         description:"实例CPU负载或利用率，单位由上报端统一"`
	MemoryAllocated float64     `json:"memoryAllocated" orm:"memory_allocated" description:"实例已分配内存容量，单位MB"`
	MemoryUsed      float64     `json:"memoryUsed"      orm:"memory_used"      description:"实例已使用内存容量，单位MB"`
	DiskIoRead      float64     `json:"diskIoRead"      orm:"disk_io_read"     description:"实例磁盘读取速率，单位KB/S"`
	DiskIoWrite     float64     `json:"diskIoWrite"     orm:"disk_io_write"    description:"实例磁盘写入速率，单位KB/S"`
	NetworkIn       float64     `json:"networkIn"       orm:"network_in"       description:"实例网络入站速率，单位KB/S"`
	NetworkOut      float64     `json:"networkOut"      orm:"network_out"      description:"实例网络出站速率，单位KB/S"`
	LiveStreams     int         `json:"liveStreams"     orm:"live_streams"     description:"实例当前承载直播流数量"`
	Sessions        int         `json:"sessions"        orm:"sessions"         description:"实例当前会话数量"`
	StartTime       *gtime.Time `json:"startTime"       orm:"start_time"       description:"实例启动时间"`
	Version         string      `json:"version"         orm:"version"          description:"实例版本号"`
	ReportTime      int64       `json:"reportTime"      orm:"report_time"      description:"上报端采样时间"`
	UpdatedAt       *gtime.Time `json:"updatedAt"       orm:"updated_at"       description:"记录更新时间"`
}
