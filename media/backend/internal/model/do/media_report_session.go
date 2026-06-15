// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"
)

// MediaReportSession is the golang structure of table media_report_session for DAO operations like Where/Data.
type MediaReportSession struct {
	g.Meta            `orm:"table:media_report_session, do:true"`
	SessionId         any         // 会话业务标识
	StreamId          any         // 会话所属流业务标识
	StreamName        any         // 会话所属流展示名称，按上报时间点冗余
	TenantId          any         // 租户标识，用于数据权限过滤和租户统计
	ClientId          any         // 客户端标识
	ClientIp          any         // 客户端IP地址
	ClientType        any         // 客户端类型
	UserName          any         // 播放用户展示名称
	ProtocolType      any         // 播放协议类型
	StartTime         *gtime.Time // 会话开始时间
	PlayDuration      any         // 播放持续时间，单位秒
	CurrentFps        any         // 当前播放帧率
	CurrentBitrate    any         // 当前播放码率，单位Kbps
	CurrentResolution any         // 当前播放分辨率
	NodeId            any         // 会话承载节点业务标识
	NodeName          any         // 会话承载节点展示名称，按上报时间点冗余
	InstanceId        any         // 会话承载实例业务标识
	InstanceName      any         // 会话承载实例展示名称，按上报时间点冗余
	LinkHops          any         // 会话链路跳点JSON数组，结构来自接口link_hops
	TotalLinkLatency  any         // 会话链路总延迟，单位毫秒
	ReportTime        any         // 上报端采样时间
	UpdatedAt         *gtime.Time // 记录更新时间
}
