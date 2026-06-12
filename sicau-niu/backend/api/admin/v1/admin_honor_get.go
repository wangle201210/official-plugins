// admin_honor_get.go defines the request and response DTOs for fetching one
// honor definition detail by ID.

package v1

import "github.com/gogf/gf/v2/frame/g"

// GetHonorReq is the request for fetching one honor definition detail.
type GetHonorReq struct {
	g.Meta `path:"/plugins/sicau-niu/admin/honors/{id}" method:"get" tags:"Sicau Niu Admin" summary:"获取荣誉定义详情" dc:"Fetch one honor definition detail by ID. Protected by host unified permission check." permission:"sicau-niu:honor:list"`
	Id     int64 `json:"id" v:"required|min:1" dc:"Honor definition ID from the path" eg:"1"`
}

// GetHonorRes is the response for fetching one honor definition detail.
type GetHonorRes struct {
	*HonorItem
}
