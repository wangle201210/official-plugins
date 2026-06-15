// This file declares media dashboard instance-list DTOs.

package v1

import "github.com/gogf/gf/v2/frame/g"

// ListDashboardInstancesReq defines the request for querying dashboard instances.
type ListDashboardInstancesReq struct {
	g.Meta  `path:"/media/dashboard/instances" method:"get" tags:"媒体数据看板" summary:"查询实例列表" dc:"查询媒体数据看板实例列表，读取采集上报实例最新投影；最多返回10000条。" permission:"media:management:query"`
	NodeId  string `json:"nodeId" dc:"按节点ID筛选实例" eg:"node-01"`
	Status  string `json:"status" dc:"按实例状态筛选，例如running、stopped、crashed" eg:"running"`
	Keyword string `json:"keyword" dc:"按实例ID、实例名称或节点名称模糊筛选" eg:"inst"`
}

// ListDashboardInstancesRes defines the dashboard instance-list response.
type ListDashboardInstancesRes struct {
	NodeInfo     *DashboardInstanceNodeInfo `json:"node_info" dc:"节点基础信息；按nodeId查询时返回对应节点信息" eg:"{}"`
	InstanceList []*DashboardInstanceItem   `json:"instance_list" dc:"实例列表" eg:"[]"`
}

// DashboardInstanceNodeInfo defines the node summary returned with instances.
type DashboardInstanceNodeInfo struct {
	NodeId   string `json:"node_id" dc:"节点ID" eg:"node-01"`
	NodeName string `json:"node_name" dc:"节点名称" eg:"怀化节点-01"`
	Region   string `json:"region" dc:"节点所属区域" eg:"怀化"`
	Status   string `json:"status" dc:"节点状态" eg:"healthy"`
}

// DashboardInstanceItem defines one dashboard instance row.
type DashboardInstanceItem struct {
	InstanceId      string  `json:"instance_id" dc:"实例唯一标识，例如容器ID" eg:"inst-001"`
	InstanceName    string  `json:"instance_name" dc:"实例名称" eg:"转码实例-01"`
	Status          string  `json:"status" dc:"实例状态，例如running、stopped、crashed" eg:"running"`
	CpuAllocated    float64 `json:"cpu_allocated" dc:"实例CPU分配核数，单位核" eg:"2"`
	CpuLoad         float64 `json:"cpu_load" dc:"实例实时CPU负载，单位核" eg:"1.3"`
	MemoryAllocated float64 `json:"memory_allocated" dc:"实例内存分配量，单位GB" eg:"8"`
	MemoryUsed      float64 `json:"memory_used" dc:"实例内存使用量，单位GB" eg:"5.2"`
	DiskIoRead      float64 `json:"disk_io_read" dc:"实例磁盘读速率，单位MB/s" eg:"120"`
	DiskIoWrite     float64 `json:"disk_io_write" dc:"实例磁盘写速率，单位MB/s" eg:"80"`
	NetworkIn       float64 `json:"network_in" dc:"实例入流量，单位Mbps" eg:"1054.6"`
	NetworkOut      float64 `json:"network_out" dc:"实例出流量，单位Mbps" eg:"3025.5"`
	LiveStreams     int     `json:"live_streams" dc:"实例内存活直播流数" eg:"8"`
	Sessions        int     `json:"sessions" dc:"实例内会话数" eg:"120"`
	StartTime       *int64  `json:"start_time" dc:"实例启动时间，Unix timestamp in milliseconds" eg:"1779245700000"`
	Version         string  `json:"version" dc:"实例版本号" eg:"v2.3.1"`
}
