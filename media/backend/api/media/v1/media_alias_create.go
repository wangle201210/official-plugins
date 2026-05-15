// This file declares media stream alias create DTOs.

package v1

import "github.com/gogf/gf/v2/frame/g"

// CreateAliasReq defines the request for creating one stream alias.
type CreateAliasReq struct {
	g.Meta     `path:"/media/stream-aliases" method:"post" tags:"媒体管理" summary:"新增流别名" dc:"新增一条流别名到真实流路径的映射。" permission:"media:management:add"`
	Alias      string `json:"alias" v:"required|length:1,255#流别名不能为空|流别名长度不能超过255个字符" dc:"流别名" eg:"camera-01"`
	AutoRemove int    `json:"autoRemove" d:"0" dc:"是否自动移除：1是，0否" eg:"0"`
	StreamPath string `json:"streamPath" v:"required|length:1,512#真实流路径不能为空|真实流路径长度不能超过512个字符" dc:"真实流路径" eg:"live/camera-01"`
	DeviceId   string `json:"deviceId" v:"required|length:1,64#设备ID不能为空|设备ID长度不能超过64个字符" dc:"设备ID" eg:"34020000001320000001"`
	ChannelId  string `json:"channelId" v:"required|length:1,64#设备通道ID不能为空|设备通道ID长度不能超过64个字符" dc:"设备通道ID" eg:"34020000001320000001"`
}

// CreateAliasRes defines the response for creating one stream alias.
type CreateAliasRes struct {
	Id int64 `json:"id" dc:"新建流别名ID" eg:"1"`
}
