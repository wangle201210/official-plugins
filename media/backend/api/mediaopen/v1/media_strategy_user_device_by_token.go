// This file declares the HotGo-compatible public media strategy endpoint.

package v1

import "github.com/gogf/gf/v2/frame/g"

// UserDeviceStrategyByTokenReq defines the HotGo-compatible request for resolving one device strategy by token.
type UserDeviceStrategyByTokenReq struct {
	g.Meta      `path:"/strategies/user-device" method:"get" tags:"内部接口" summary:"通过铁塔 token 和设备ID查询媒体策略" dc:"兼容 HotGo Token+DeviceId 策略查询接口，使用铁塔 token 换取租户身份并校验租户设备权限。" access:"public"`
	InnerApiKey string `json:"X-Inner-Api-Key" in:"header" dc:"内部接口API Key；默认值media；显式配置innerapi.apiKey为空时可不传" eg:"media"`
	Token       string `json:"token" v:"required#Token不能为空" dc:"铁塔用户 token，直接传 token 原值" eg:"token-value"`
	DeviceId    string `json:"deviceId" v:"required#设备ID不能为空" dc:"设备国标编号" eg:"34020000001320000001"`
	NodeId      string `json:"nodeId" v:"required#节点ID不能为空" dc:"节点ID，用于按租户和节点检查流数量限制" eg:"1"`
}

// UserDeviceStrategyByTokenRes defines the HotGo-compatible response for resolving one device strategy by token.
type UserDeviceStrategyByTokenRes struct {
	UserInfo *TietaUserInfo `json:"userInfo" dc:"铁塔用户信息" eg:"{}"`
	Strategy *StrategyInfo  `json:"strategy" dc:"策略详情；无设备权限或无匹配策略时为空" eg:"{}"`
}

// TietaUserInfo defines the HotGo-compatible Tieta identity projection returned by token validation.
type TietaUserInfo struct {
	Id           int64  `json:"id" dc:"用户ID" eg:"13"`
	CustomerName string `json:"customerName" dc:"租户名" eg:"公安"`
	Phone        string `json:"phone" dc:"手机号码" eg:"18213268117"`
}

// StrategyInfo defines the HotGo-compatible strategy payload.
type StrategyInfo struct {
	Id              uint64 `json:"id" dc:"策略ID" eg:"1"`
	Name            string `json:"name" dc:"策略名称" eg:"默认直播策略"`
	StrategyContent string `json:"strategyContent" dc:"YAML格式策略内容" eg:"record: true"`
}
