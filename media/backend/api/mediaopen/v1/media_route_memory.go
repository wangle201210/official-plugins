// This file declares HotGo-compatible public media route-memory endpoints.

package v1

import "github.com/gogf/gf/v2/frame/g"

// SetRouteDataReq defines the HotGo-compatible request for writing route data.
type SetRouteDataReq struct {
	g.Meta      `path:"/route/set" method:"post" tags:"路由记忆" summary:"写入路由数据" dc:"兼容 HotGo 路由记忆接口，按设备编号和通道编号保存路由数据。" access:"public"`
	InnerApiKey string `json:"X-Inner-Api-Key" in:"header" dc:"内部接口API Key；默认值media；显式配置innerapi.apiKey为空时可不传" eg:"media"`
	DeviceCode  string `json:"deviceCode" v:"required#设备编号不能为空" dc:"设备编号" eg:"34020000001320000001"`
	ChannelCode string `json:"channelCode" v:"required#通道编号不能为空" dc:"通道编号" eg:"34020000001320000001"`
	Data        string `json:"data" v:"required#数据不能为空" dc:"路由数据" eg:"node-a"`
}

// SetRouteDataRes defines the HotGo-compatible response for writing route data.
type SetRouteDataRes struct{}

// GetRouteDataReq defines the HotGo-compatible request for reading route data.
type GetRouteDataReq struct {
	g.Meta      `path:"/route/get" method:"post" tags:"路由记忆" summary:"读取路由数据" dc:"兼容 HotGo 路由记忆接口，按设备编号和通道编号读取路由数据。" access:"public"`
	InnerApiKey string `json:"X-Inner-Api-Key" in:"header" dc:"内部接口API Key；默认值media；显式配置innerapi.apiKey为空时可不传" eg:"media"`
	DeviceCode  string `json:"deviceCode" v:"required#设备编号不能为空" dc:"设备编号" eg:"34020000001320000001"`
	ChannelCode string `json:"channelCode" v:"required#通道编号不能为空" dc:"通道编号" eg:"34020000001320000001"`
}

// GetRouteDataRes defines the HotGo-compatible response for reading route data.
type GetRouteDataRes struct {
	Data string `json:"data" dc:"路由数据" eg:"node-a"`
}

// DelRouteDataReq defines the HotGo-compatible request for deleting route data.
type DelRouteDataReq struct {
	g.Meta      `path:"/route/del" method:"post" tags:"路由记忆" summary:"删除路由数据" dc:"兼容 HotGo 路由记忆接口，按设备编号和通道编号删除路由数据。" access:"public"`
	InnerApiKey string `json:"X-Inner-Api-Key" in:"header" dc:"内部接口API Key；默认值media；显式配置innerapi.apiKey为空时可不传" eg:"media"`
	DeviceCode  string `json:"deviceCode" v:"required#设备编号不能为空" dc:"设备编号" eg:"34020000001320000001"`
	ChannelCode string `json:"channelCode" v:"required#通道编号不能为空" dc:"通道编号" eg:"34020000001320000001"`
}

// DelRouteDataRes defines the HotGo-compatible response for deleting route data.
type DelRouteDataRes struct{}
