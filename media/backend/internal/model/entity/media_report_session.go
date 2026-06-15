// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

import (
	"github.com/gogf/gf/v2/os/gtime"
)

// MediaReportSession is the golang structure for table media_report_session.
type MediaReportSession struct {
	SessionId         string      `json:"sessionId"         orm:"session_id"         description:"会话业务标识"`
	StreamId          string      `json:"streamId"          orm:"stream_id"          description:"会话所属流业务标识"`
	StreamName        string      `json:"streamName"        orm:"stream_name"        description:"会话所属流展示名称，按上报时间点冗余"`
	TenantId          string      `json:"tenantId"          orm:"tenant_id"          description:"租户标识，用于数据权限过滤和租户统计"`
	ClientId          string      `json:"clientId"          orm:"client_id"          description:"客户端标识"`
	ClientIp          string      `json:"clientIp"          orm:"client_ip"          description:"客户端IP地址"`
	ClientType        string      `json:"clientType"        orm:"client_type"        description:"客户端类型"`
	UserName          string      `json:"userName"          orm:"user_name"          description:"播放用户展示名称"`
	ProtocolType      string      `json:"protocolType"      orm:"protocol_type"      description:"播放协议类型"`
	StartTime         *gtime.Time `json:"startTime"         orm:"start_time"         description:"会话开始时间"`
	PlayDuration      int         `json:"playDuration"      orm:"play_duration"      description:"播放持续时间，单位秒"`
	CurrentFps        float64     `json:"currentFps"        orm:"current_fps"        description:"当前播放帧率"`
	CurrentBitrate    int         `json:"currentBitrate"    orm:"current_bitrate"    description:"当前播放码率，单位Kbps"`
	CurrentResolution string      `json:"currentResolution" orm:"current_resolution" description:"当前播放分辨率"`
	NodeId            string      `json:"nodeId"            orm:"node_id"            description:"会话承载节点业务标识"`
	NodeName          string      `json:"nodeName"          orm:"node_name"          description:"会话承载节点展示名称，按上报时间点冗余"`
	InstanceId        string      `json:"instanceId"        orm:"instance_id"        description:"会话承载实例业务标识"`
	InstanceName      string      `json:"instanceName"      orm:"instance_name"      description:"会话承载实例展示名称，按上报时间点冗余"`
	LinkHops          string      `json:"linkHops"          orm:"link_hops"          description:"会话链路跳点JSON数组，结构来自接口link_hops"`
	TotalLinkLatency  int         `json:"totalLinkLatency"  orm:"total_link_latency" description:"会话链路总延迟，单位毫秒"`
	ReportTime        int64       `json:"reportTime"        orm:"report_time"        description:"上报端采样时间"`
	UpdatedAt         *gtime.Time `json:"updatedAt"         orm:"updated_at"         description:"记录更新时间"`
}
