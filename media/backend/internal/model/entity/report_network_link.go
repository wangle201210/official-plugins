// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

// ReportNetworkLink is the golang structure for table report_network_link.
type ReportNetworkLink struct {
	NodeId        string  `json:"nodeId"        orm:"node_id"        description:"上报所属节点业务标识"`
	MachineId     string  `json:"machineId"     orm:"machine_id"     description:"上报来源机器或实例标识"`
	SourceIp      string  `json:"sourceIp"      orm:"source_ip"      description:"原始探测源地址"`
	DestinationIp string  `json:"destinationIp" orm:"destination_ip" description:"原始探测目的地址"`
	Rtt           int     `json:"rtt"           orm:"rtt"            description:"原始往返延迟，单位毫秒，保留探测失败固定值"`
	Jitter        int     `json:"jitter"        orm:"jitter"         description:"原始抖动，单位毫秒，保留探测失败固定值"`
	PacketLoss    float64 `json:"packetLoss"    orm:"packet_loss"    description:"原始丢包比例0至1，当前上游口径为HTTP探测失败比例"`
	StatusCode    int     `json:"statusCode"    orm:"status_code"    description:"原始探测HTTP状态码"`
	Reachable     *bool   `json:"reachable"     orm:"reachable"      description:"原始extra.reachable布尔状态，未提供或无法解析时为空"`
	ReportTime    int64   `json:"reportTime"    orm:"report_time"    description:"上报端采样时间，Unix毫秒时间戳"`
}
