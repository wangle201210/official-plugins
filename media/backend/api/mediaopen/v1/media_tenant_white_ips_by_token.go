// This file declares the public mediaopen tenant whitelist IP query endpoint.

package v1

import "github.com/gogf/gf/v2/frame/g"

// TenantWhiteIPsByTokenReq defines the request for querying enabled tenant whitelist IPs by user token.
type TenantWhiteIPsByTokenReq struct {
	g.Meta      `path:"/tenant-whites/ips" method:"post" tags:"媒体鉴权" summary:"通过用户 token 查询租户 IP 白名单" dc:"通过用户 token 解析所属租户，并返回该租户下已启用的 IP 白名单数组。"`
	InnerApiKey string `json:"X-Inner-Api-Key" in:"header" dc:"内部接口API Key；默认值media；显式配置innerapi.apiKey为空时可不传" eg:"media"`
	Token       string `json:"token" v:"required#用户 token 不能为空" dc:"用户 token 请求参数，必填" eg:"token-value"`
}

// TenantWhiteIPsByTokenRes defines the enabled whitelist IP array returned without pagination.
type TenantWhiteIPsByTokenRes []string
