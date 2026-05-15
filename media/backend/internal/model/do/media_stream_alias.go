// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"
)

// MediaStreamAlias is the golang structure of table media_stream_alias for DAO operations like Where/Data.
type MediaStreamAlias struct {
	g.Meta     `orm:"table:media_stream_alias, do:true"`
	Id         any         // ID（自增，无符号）
	Alias      any         // 流别名
	AutoRemove any         // 是否自动移除（0否1是）
	StreamPath any         // 真实流路径
	DeviceId   any         // 设备id（对应device_code）
	ChannelId  any         // 设备通道id（对应channel_code）
	CreateTime *gtime.Time // 创建时间
}
