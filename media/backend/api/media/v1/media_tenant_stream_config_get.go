// This file declares media tenant stream config detail DTOs.

package v1

import "github.com/gogf/gf/v2/frame/g"

// GetTenantStreamConfigReq defines the request for querying one tenant stream config.
type GetTenantStreamConfigReq struct {
	g.Meta   `path:"/media/tenant-stream-configs/{tenantId}/nodes/{nodeNum}" method:"get" tags:"租户流配置" summary:"查询租户流配置详情" dc:"按租户ID和节点编号查询租户流配置详情。" permission:"media:management:query"`
	TenantId string `json:"tenantId" v:"required|length:1,64#租户ID不能为空|租户ID长度不能超过64个字符" dc:"租户ID" eg:"tenant-a"`
	NodeNum  int    `json:"nodeNum" v:"min:0|max:255#节点编号不能小于0|节点编号不能大于255" dc:"节点编号" eg:"1"`
}

// GetTenantStreamConfigRes defines the response for querying one tenant stream config.
type GetTenantStreamConfigRes = TenantStreamConfigListItem
