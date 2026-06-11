// This file declares media tenant stream config delete DTOs.

package v1

import "github.com/gogf/gf/v2/frame/g"

// DeleteTenantStreamConfigReq defines the request for deleting one tenant stream config.
type DeleteTenantStreamConfigReq struct {
	g.Meta   `path:"/media/tenant-stream-configs/{tenantId}/nodes/{nodeNum}" method:"delete" tags:"租户流配置" summary:"删除租户流配置" dc:"按租户ID和节点编号删除租户流配置。" permission:"media:management:remove"`
	TenantId string `json:"tenantId" v:"required|length:1,64#租户ID不能为空|租户ID长度不能超过64个字符" dc:"租户ID" eg:"tenant-a"`
	NodeNum  int    `json:"nodeNum" v:"min:0|max:255#节点编号不能小于0|节点编号不能大于255" dc:"节点编号" eg:"1"`
}

// DeleteTenantStreamConfigRes defines the response for deleting one tenant stream config.
type DeleteTenantStreamConfigRes struct {
	TenantId string `json:"tenantId" dc:"租户ID" eg:"tenant-a"`
	NodeNum  int    `json:"nodeNum" dc:"节点编号" eg:"1"`
}
