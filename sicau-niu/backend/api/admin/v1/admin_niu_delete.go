// admin_niu_delete.go defines the request and response DTOs for deleting one
// cattle and its main card.

package v1

import "github.com/gogf/gf/v2/frame/g"

// DeleteNiuReq is the request for deleting one cattle.
type DeleteNiuReq struct {
	g.Meta `path:"/plugins/sicau-niu/admin/niu/{id}" method:"delete" tags:"Sicau Niu Admin" summary:"Delete cattle" dc:"Soft-delete a cattle and cascade soft-delete its unique main card in one transaction, so no dangling card remains. Both are recoverable. Protected by host unified permission check." permission:"sicau-niu:niu:delete"`
	Id     int64 `json:"id" v:"required|min:1" dc:"Cattle ID from the path" eg:"1"`
}

// DeleteNiuRes is the response for deleting one cattle. It is intentionally
// empty; success is conveyed by the absence of an error.
type DeleteNiuRes struct{}
