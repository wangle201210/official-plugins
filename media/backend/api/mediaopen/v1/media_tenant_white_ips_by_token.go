// This file declares the public mediaopen tenant whitelist IP query endpoint.

package v1

import "github.com/gogf/gf/v2/frame/g"

// TenantWhiteIPsByTokenReq defines the request for querying enabled tenant whitelist IPs by user token.
type TenantWhiteIPsByTokenReq struct {
	g.Meta      `path:"/tenant-whites/ips" method:"post" tags:"媒体鉴权" summary:"通过用户 token 查询租户 IP 白名单" dc:"通过用户 token 解析所属租户，并返回该租户ID及已启用的 IP 白名单数组。" access:"public"`
	InnerApiKey string `json:"X-Inner-Api-Key" in:"header" dc:"内部接口API Key；默认值media；显式配置innerapi.apiKey为空时可不传" eg:"media"`
	Token       string `json:"token" v:"required#用户 token 不能为空" dc:"用户 token 请求参数，必填；直接传 token 原值" eg:"token-value"`
}

// TenantWhiteIPsByTokenRes defines the tenant-scoped whitelist IP response without pagination.
type TenantWhiteIPsByTokenRes struct {
	TenantId string   `json:"tenantId" dc:"租户ID，客户端可用于租户维度缓存" eg:"tenant-a"`
	Ips      []string `json:"ips" dc:"已启用的租户白名单 IP 数组" eg:"[]"`
}
