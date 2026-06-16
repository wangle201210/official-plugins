// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

import (
	"github.com/gogf/gf/v2/os/gtime"
)

// MediaReportNode is the golang structure for table media_report_node.
type MediaReportNode struct {
	NodeId          string      `json:"nodeId"          orm:"node_id"          description:"上报节点业务标识，不依赖配置表外键"`
	NodeName        string      `json:"nodeName"        orm:"node_name"        description:"节点展示名称，按上报时间点冗余"`
	Region          string      `json:"region"          orm:"region"           description:"节点所属区域或机房名称"`
	ParentNodeId    string      `json:"parentNodeId"    orm:"parent_node_id"   description:"父节点业务标识，根节点使用0"`
	Status          string      `json:"status"          orm:"status"           description:"节点运行状态，保存上报原始枚举值"`
	TotalNodes      int         `json:"totalNodes"      orm:"total_nodes"      description:"当前节点视角下的节点总数"`
	AliveNodes      int         `json:"aliveNodes"      orm:"alive_nodes"      description:"当前节点视角下的在线节点数"`
	CpuAllocated    float64     `json:"cpuAllocated"    orm:"cpu_allocated"    description:"已分配CPU容量，单位由上报端统一"`
	CpuLoad         float64     `json:"cpuLoad"         orm:"cpu_load"         description:"CPU负载或利用率，单位由上报端统一"`
	MemoryAllocated float64     `json:"memoryAllocated" orm:"memory_allocated" description:"已分配内存容量，单位MB"`
	MemoryUsed      float64     `json:"memoryUsed"      orm:"memory_used"      description:"已使用内存容量，单位MB"`
	DiskIoRead      float64     `json:"diskIoRead"      orm:"disk_io_read"     description:"磁盘读取速率，单位KB/S"`
	DiskIoWrite     float64     `json:"diskIoWrite"     orm:"disk_io_write"    description:"磁盘写入速率，单位KB/S"`
	NetworkIn       float64     `json:"networkIn"       orm:"network_in"       description:"网络入站速率，单位KB/S"`
	NetworkOut      float64     `json:"networkOut"      orm:"network_out"      description:"网络出站速率，单位KB/S"`
	LiveStreams     int         `json:"liveStreams"     orm:"live_streams"     description:"当前直播流数量"`
	Sessions        int         `json:"sessions"        orm:"sessions"         description:"当前会话数量"`
	AvgDelay        int         `json:"avgDelay"        orm:"avg_delay"        description:"节点平均延迟，单位毫秒"`
	LastHeartbeat   *gtime.Time `json:"lastHeartbeat"   orm:"last_heartbeat"   description:"节点最后心跳时间"`
	NodeLatencyMap  string      `json:"nodeLatencyMap"  orm:"node_latency_map" description:"节点到其他节点的延迟矩阵JSON，结构来自接口node_latency_map"`
	ReportTime      int64       `json:"reportTime"      orm:"report_time"      description:"上报端采样时间"`
	UpdatedAt       *gtime.Time `json:"updatedAt"       orm:"updated_at"       description:"记录更新时间"`
}
