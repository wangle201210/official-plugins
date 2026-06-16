// This file declares media dashboard node overview DTOs.

package v1

import "github.com/gogf/gf/v2/frame/g"

// GetDashboardNodeOverviewReq defines the request for querying the dashboard node tree.
type GetDashboardNodeOverviewReq struct {
	g.Meta       `path:"/media/dashboard/nodes/overview" method:"get" tags:"媒体数据看板" summary:"查询节点总览" dc:"查询媒体数据看板节点总览树，读取采集上报最新投影并按父子节点组装。" permission:"media:management:query"`
	RootNodeId   string `json:"rootNodeId" dc:"根节点ID；为空时返回全部根节点，传入时返回该节点及其子树" eg:"0"`
	IncludeEmpty bool   `json:"includeEmpty" dc:"是否在没有根节点匹配时返回空节点；默认false" eg:"false"`
}

// GetDashboardNodeOverviewRes defines the dashboard node overview response.
type GetDashboardNodeOverviewRes = DashboardNodeOverviewItem

// DashboardNodeOverviewItem defines one dashboard node overview tree item.
type DashboardNodeOverviewItem struct {
	NodeId          string                       `json:"node_id" dc:"当前节点唯一标识" eg:"node-01"`
	NodeName        string                       `json:"node_name" dc:"当前节点名称" eg:"省节点-01"`
	Region          string                       `json:"region" dc:"当前所属区域" eg:"湖南省"`
	Status          string                       `json:"status" dc:"当前节点状态，例如healthy、warning、error、offLine" eg:"healthy"`
	ParentNodeId    string                       `json:"parent_node_id" dc:"父节点标识，根节点为0" eg:"0"`
	TotalNodes      int                          `json:"total_nodes" dc:"总节点数" eg:"8"`
	AliveNodes      int                          `json:"alive_nodes" dc:"存活节点数" eg:"7"`
	CpuAllocated    float64                      `json:"cpu_allocated" dc:"当前节点CPU分配核数，单位核" eg:"64"`
	CpuLoad         float64                      `json:"cpu_load" dc:"当前节点CPU负载，单位核" eg:"42.5"`
	MemoryAllocated float64                      `json:"memory_allocated" dc:"当前节点内存分配量，单位MB" eg:"262144"`
	MemoryUsed      float64                      `json:"memory_used" dc:"当前节点内存使用量，单位MB" eg:"186675.2"`
	DiskIoRead      float64                      `json:"disk_io_read" dc:"当前节点磁盘读速率，单位KB/S" eg:"120"`
	DiskIoWrite     float64                      `json:"disk_io_write" dc:"当前节点磁盘写速率，单位KB/S" eg:"80"`
	NetworkIn       float64                      `json:"network_in" dc:"当前节点入流量，单位KB/S" eg:"1054.6"`
	NetworkOut      float64                      `json:"network_out" dc:"当前节点出流量，单位KB/S" eg:"3025.5"`
	LiveStreams     int                          `json:"live_streams" dc:"当前节点全局存活直播流数" eg:"156"`
	Sessions        int                          `json:"sessions" dc:"当前节点全局会话数" eg:"2340"`
	AvgDelay        int                          `json:"avg_delay" dc:"当前节点全局平均流延迟，单位毫秒" eg:"280"`
	LastHeartbeat   *int64                       `json:"last_heartbeat" dc:"当前节点最后心跳时间，Unix timestamp in milliseconds" eg:"1779268200000"`
	ReportTime      int64                        `json:"report_time" dc:"消息上报时间戳；保持上报端原始秒或毫秒精度" eg:"1780923543"`
	NodeLatencyMap  map[string]int               `json:"node_latency_map" dc:"当前节点到其他节点的延迟，单位毫秒" eg:"{\"node-02\":18}"`
	ChildNodes      []*DashboardNodeOverviewItem `json:"child_nodes" dc:"子节点信息列表，结构同父层" eg:"[]"`
}
