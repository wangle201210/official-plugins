// This file declares the public mediaopen stream-alias lookup endpoint.

package v1

import "github.com/gogf/gf/v2/frame/g"

// GetStreamAliasByAliasReq defines the request for reading one stream alias config by alias.
type GetStreamAliasByAliasReq struct {
	g.Meta      `path:"/stream-aliases/by-alias" method:"get" tags:"媒体配置" summary:"通过流别名查询别名配置" dc:"通过流别名查询对应的真实流路径、设备和通道配置。" access:"public"`
	InnerApiKey string `json:"X-Inner-Api-Key" in:"header" dc:"内部接口API Key；默认值media；显式配置innerapi.apiKey为空时可不传" eg:"media"`
	Alias       string `json:"alias" v:"required#流别名不能为空" dc:"流别名" eg:"camera-01"`
}

// GetStreamAliasByAliasRes defines one stream alias config response.
type GetStreamAliasByAliasRes struct {
	Id         int64  `json:"id" dc:"ID" eg:"1"`
	Alias      string `json:"alias" dc:"流别名" eg:"camera-01"`
	AutoRemove int    `json:"autoRemove" dc:"是否自动移除：1是，0否" eg:"0"`
	StreamPath string `json:"streamPath" dc:"真实流路径" eg:"live/camera-01"`
	DeviceId   string `json:"deviceId" dc:"设备ID" eg:"34020000001320000001"`
	ChannelId  string `json:"channelId" dc:"设备通道ID" eg:"34020000001320000001"`
	CreateTime *int64 `json:"createTime" dc:"创建时间，Unix timestamp in milliseconds" eg:"1778733600000"`
}
