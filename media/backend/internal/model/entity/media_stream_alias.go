// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

import (
	"github.com/gogf/gf/v2/os/gtime"
)

// MediaStreamAlias is the golang structure for table media_stream_alias.
type MediaStreamAlias struct {
	Id         int64       `json:"id"         orm:"id"          description:"ID（自增，无符号）"`
	Alias      string      `json:"alias"      orm:"alias"       description:"流别名"`
	AutoRemove int         `json:"autoRemove" orm:"auto_remove" description:"是否自动移除（0否1是）"`
	StreamPath string      `json:"streamPath" orm:"stream_path" description:"真实流路径"`
	DeviceId   string      `json:"deviceId"   orm:"device_id"   description:"设备id（对应device_code）"`
	ChannelId  string      `json:"channelId"  orm:"channel_id"  description:"设备通道id（对应channel_code）"`
	CreateTime *gtime.Time `json:"createTime" orm:"create_time" description:"创建时间"`
}
