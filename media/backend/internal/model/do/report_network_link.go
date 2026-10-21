// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
)

// ReportNetworkLink is the golang structure of table plugin_media_report_network_link for DAO operations like Where/Data.
type ReportNetworkLink struct {
	g.Meta        `orm:"table:plugin_media_report_network_link, do:true"`
	NodeId        any   // 上报所属节点业务标识
	MachineId     any   // 上报来源机器或实例标识
	SourceIp      any   // 原始探测源地址
	DestinationIp any   // 原始探测目的地址
	Rtt           any   // 原始往返延迟，单位毫秒，保留探测失败固定值
	Jitter        any   // 原始抖动，单位毫秒，保留探测失败固定值
	PacketLoss    any   // 原始丢包比例0至1，当前上游口径为HTTP探测失败比例
	StatusCode    any   // 原始探测HTTP状态码
	Reachable     *bool // 原始extra.reachable布尔状态，未提供或无法解析时为空
	ReportTime    any   // 上报端采样时间，Unix毫秒时间戳
}
