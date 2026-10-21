// This file declares media dashboard node overview DTOs.

package v1

import "github.com/gogf/gf/v2/frame/g"

// GetDashboardNodeOverviewReq defines the request for querying the dashboard node tree.
type GetDashboardNodeOverviewReq struct {
	g.Meta       `path:"/media/dashboard/nodes/overview" method:"get" tags:"媒体数据看板" summary:"查询节点总览" dc:"查询媒体节点总览树、实例资源汇总和原始网络链路指标；节点最多10000个，当前子树链路明细最多10000条，不计算质量评分或分级。" permission:"media:management:query"`
	RootNodeId   string `json:"rootNodeId" dc:"根节点ID；为空时返回全部根节点，传入时返回该节点及其子树" eg:"0"`
	IncludeEmpty bool   `json:"includeEmpty" dc:"是否在没有根节点匹配时返回空节点；默认false" eg:"false"`
}

// GetDashboardNodeOverviewRes defines the dashboard node overview response.
type GetDashboardNodeOverviewRes = DashboardNodeOverviewItem

// DashboardNodeOverviewItem defines one dashboard node overview tree item.
type DashboardNodeOverviewItem struct {
	NodeId              string                       `json:"node_id" dc:"当前节点唯一标识" eg:"node-01"`
	NodeName            string                       `json:"node_name" dc:"当前节点名称" eg:"省节点-01"`
	Region              string                       `json:"region" dc:"当前所属区域" eg:"湖南省"`
	Status              string                       `json:"status" dc:"当前节点状态，例如healthy、warning、error、offLine" eg:"healthy"`
	ParentNodeId        string                       `json:"parent_node_id" dc:"父节点标识，根节点为0" eg:"0"`
	TotalNodes          int                          `json:"total_nodes" dc:"总节点数" eg:"8"`
	AliveNodes          int                          `json:"alive_nodes" dc:"存活节点数" eg:"7"`
	CpuAllocated        float64                      `json:"cpu_allocated" dc:"当前节点CPU分配核数，单位核" eg:"64"`
	CpuLoad             float64                      `json:"cpu_load" dc:"当前节点CPU负载，单位核" eg:"42.5"`
	MemoryAllocated     float64                      `json:"memory_allocated" dc:"当前节点内存分配量，单位MB" eg:"262144"`
	MemoryUsed          float64                      `json:"memory_used" dc:"当前节点内存使用量，单位MB" eg:"186675.2"`
	DiskIoRead          float64                      `json:"disk_io_read" dc:"当前节点磁盘读速率，单位KB/S" eg:"120"`
	DiskIoWrite         float64                      `json:"disk_io_write" dc:"当前节点磁盘写速率，单位KB/S" eg:"80"`
	NetworkIn           float64                      `json:"network_in" dc:"当前节点入流量，单位KB/S" eg:"1054.6"`
	NetworkOut          float64                      `json:"network_out" dc:"当前节点出流量，单位KB/S" eg:"3025.5"`
	LiveStreams         int                          `json:"live_streams" dc:"当前节点全局存活直播流数" eg:"156"`
	Sessions            int                          `json:"sessions" dc:"当前节点全局会话数" eg:"2340"`
	AvgDelay            int                          `json:"avg_delay" dc:"当前节点全局平均流延迟，单位毫秒" eg:"280"`
	LastHeartbeat       *int64                       `json:"last_heartbeat" dc:"当前节点最后心跳时间，Unix timestamp in milliseconds" eg:"1779268200000"`
	ReportTime          int64                        `json:"report_time" dc:"消息上报时间戳；保持上报端原始秒或毫秒精度" eg:"1780923543"`
	NodeLatencyMap      map[string]int               `json:"node_latency_map" dc:"当前节点到各目的地址的最近延迟摘要，键为上报的destination_ip，单位毫秒；来源实例明细见network_links" eg:"{\"192.0.2.55\":18}"`
	NetworkLinks        []*DashboardNetworkLink      `json:"network_links" dc:"当前节点按来源实例、源地址和目的地址区分的原始链路指标；没有上报时为空数组" eg:"[]"`
	NetworkLinksLimited bool                         `json:"network_links_limited" dc:"仅响应根节点表示当前返回子树的链路明细是否超过10000条而被截断；子节点为false" eg:"false"`
	InstanceResources   *DashboardNodeResources      `json:"instance_resources" dc:"当前节点直属实例的资源汇总，不包含子节点；无实例时所有值为0，供读取途经节点资源使用，不计算综合负载" eg:"{}"`
	ChildNodes          []*DashboardNodeOverviewItem `json:"child_nodes" dc:"子节点信息列表，结构同父层" eg:"[]"`
}

// DashboardNodeResources describes direct instance totals without child-node double counting.
type DashboardNodeResources struct {
	InstanceCount   int     `json:"instance_count" dc:"当前节点直属实例数量" eg:"2"`
	CpuAllocated    float64 `json:"cpu_allocated" dc:"直属实例CPU分配总核数" eg:"8"`
	CpuLoad         float64 `json:"cpu_load" dc:"直属实例CPU已用总核数" eg:"2.5"`
	MemoryAllocated float64 `json:"memory_allocated" dc:"直属实例已分配内存总量，单位MB" eg:"8192"`
	MemoryUsed      float64 `json:"memory_used" dc:"直属实例已用内存总量，单位MB" eg:"4096"`
	DiskIoRead      float64 `json:"disk_io_read" dc:"直属实例磁盘读取总速率，单位KB/S" eg:"120"`
	DiskIoWrite     float64 `json:"disk_io_write" dc:"直属实例磁盘写入总速率，单位KB/S" eg:"80"`
	NetworkIn       float64 `json:"network_in" dc:"直属实例网络入站总速率，单位KB/S" eg:"1024"`
	NetworkOut      float64 `json:"network_out" dc:"直属实例网络出站总速率，单位KB/S" eg:"2048"`
}

// DashboardNetworkLink exposes one latest source-instance-to-destination probe without scoring.
type DashboardNetworkLink struct {
	MachineId     string  `json:"machine_id" dc:"上报来源机器或实例标识；保留NetworkMetric.machine_id原值" eg:"gateway-01"`
	SourceIp      string  `json:"source_ip" dc:"探测源地址" eg:"192.0.2.10"`
	DestinationIp string  `json:"destination_ip" dc:"探测目的地址；可能为IP或上报端无法解析时保留的主机名，不等同于节点ID" eg:"192.0.2.55"`
	Rtt           int     `json:"rtt" dc:"上报原始延迟，单位毫秒；探测失败时保留上游固定值，须结合reachable判断" eg:"18"`
	Jitter        int     `json:"jitter" dc:"上报原始抖动，单位毫秒；探测失败时保留上游固定值" eg:"3"`
	PacketLoss    float64 `json:"packet_loss" dc:"上报原始丢包比例，0表示无丢包、1表示全部失败；当前net-flux采集器实际按HTTP探测失败次数除以总次数计算" eg:"0.02"`
	StatusCode    int     `json:"status_code" dc:"上报的探测状态码；当前HTTP采集器使用HTTP状态码，0表示未取得有效状态码" eg:"200"`
	Reachable     *bool   `json:"reachable" dc:"上报extra.reachable：true可达、false不可达；缺省或无法解析时为null，不根据延迟或状态码推断" eg:"true"`
	ReportTime    int64   `json:"report_time" dc:"本链路采样时间，Unix timestamp in milliseconds" eg:"1791784485000"`
}
