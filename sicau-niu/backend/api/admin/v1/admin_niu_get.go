// admin_niu_get.go defines the request and response DTOs for fetching one
// cattle detail by ID.

package v1

import "github.com/gogf/gf/v2/frame/g"

// GetNiuReq is the request for fetching one cattle detail.
type GetNiuReq struct {
	g.Meta `path:"/plugins/sicau-niu/admin/niu/{id}" method:"get" tags:"Sicau Niu Admin" summary:"获取牛只详情" dc:"Fetch one cattle detail by ID, including the batch-assembled college name and main-card binding flag. Protected by host unified permission check." permission:"sicau-niu:niu:list"`
	Id     int64 `json:"id" v:"required|min:1" dc:"Cattle ID from the path" eg:"1"`
}

// GetNiuRes is the response for fetching one cattle detail.
type GetNiuRes struct {
	*NiuItem
}
