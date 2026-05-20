// This file declares the HotGo-compatible public media strategy endpoint.

package v1

import "github.com/gogf/gf/v2/frame/g"

// UserDeviceStrategyByTokenReq defines the HotGo-compatible request for resolving one device strategy by token.
type UserDeviceStrategyByTokenReq struct {
	g.Meta      `path:"/strategy/userDeviceStrategyByToken" method:"post" tags:"媒体鉴权" summary:"通过铁塔 token 和设备ID查询媒体策略" dc:"兼容 HotGo Token+DeviceId 策略查询接口，使用铁塔 token 换取租户身份并校验租户设备权限。"`
	InnerApiKey string `json:"X-Inner-Api-Key" in:"header" dc:"内部接口API Key；默认值media；显式配置innerapi.apiKey为空时可不传" eg:"media"`
	Token       string `json:"token" v:"required#Token不能为空" dc:"铁塔用户Token" eg:"Bearer token-value"`
	DeviceId    string `json:"deviceId" v:"required#设备ID不能为空" dc:"设备国标编号" eg:"34020000001320000001"`
}

// UserDeviceStrategyByTokenRes defines the HotGo-compatible response for resolving one device strategy by token.
type UserDeviceStrategyByTokenRes struct {
	UserInfo   *TietaUserInfo `json:"userInfo" dc:"铁塔用户信息"`
	HasAccess  bool           `json:"hasAccess" dc:"是否有权限访问该设备" eg:"true"`
	StrategyId uint64         `json:"strategyId" dc:"策略ID" eg:"1"`
	Strategy   *StrategyInfo  `json:"strategy" dc:"策略详情"`
}

// TietaUserInfo defines the HotGo-compatible Tieta identity projection returned by token validation.
type TietaUserInfo struct {
	Id           int64  `json:"id" dc:"用户ID" eg:"13"`
	DeptId       int64  `json:"deptId" dc:"部门ID" eg:"1"`
	Username     string `json:"username" dc:"用户名" eg:"wj530"`
	RealName     string `json:"realName" dc:"姓名" eg:"王杰"`
	Mobile       string `json:"mobile" dc:"手机号码" eg:"18213268117"`
	UserType     string `json:"userType" dc:"用户类型" eg:"tenant"`
	CustomerCode string `json:"customerCode" dc:"客户编码" eg:"customer-a"`
	TenantID     string `json:"tenant_id" dc:"租户ID" eg:"tenant-a"`
	DeptName     string `json:"deptName" dc:"部门名称" eg:"运营中心"`
	RegionCode   int64  `json:"regionCode" dc:"区域编码" eg:"510100"`
	OrgId        int64  `json:"orgId" dc:"组织ID" eg:"1"`
	Enable       bool   `json:"enable" dc:"是否启用" eg:"true"`
}

// StrategyInfo defines the HotGo-compatible strategy payload.
type StrategyInfo struct {
	Id              uint64 `json:"id" dc:"策略ID" eg:"1"`
	Name            string `json:"name" dc:"策略名称" eg:"默认直播策略"`
	StrategyContent string `json:"strategyContent" dc:"YAML格式策略内容" eg:"record: true"`
}
