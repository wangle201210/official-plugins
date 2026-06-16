// This file declares the internal media strategy resolution API.

package v1

import "github.com/gogf/gf/v2/frame/g"

// ResolveStrategyReq defines the internal request for resolving one effective strategy.
type ResolveStrategyReq struct {
	g.Meta      `path:"/strategies/resolve" method:"get" tags:"内部接口" summary:"解析媒体策略" dc:"内部服务按租户ID和设备ID解析当前生效媒体策略，用于跨集群服务读取策略配置。" access:"public"`
	InnerApiKey string `json:"X-Inner-Api-Key" in:"header" dc:"内部接口API Key；默认值media；显式配置innerapi.apiKey为空时可不传" eg:"media"`
	TenantId    string `json:"tenantId" dc:"媒体租户ID；为空时跳过租户策略匹配" eg:"tenant-a"`
	DeviceId    string `json:"deviceId" dc:"设备国标ID；为空时跳过设备策略匹配" eg:"34020000001320000001"`
}

// ResolveStrategyRes defines the internal response for resolving one effective strategy.
type ResolveStrategyRes struct {
	Matched      bool   `json:"matched" dc:"是否匹配到策略" eg:"true"`
	Source       string `json:"source" dc:"策略来源：tenantDevice、device、tenant、global或none" eg:"device"`
	SourceLabel  string `json:"sourceLabel" dc:"策略来源说明" eg:"设备策略"`
	StrategyId   int64  `json:"strategyId" dc:"策略ID" eg:"1"`
	StrategyName string `json:"strategyName" dc:"策略名称" eg:"默认直播策略"`
	Strategy     string `json:"strategy" dc:"YAML格式策略内容" eg:"record: true"`
}
