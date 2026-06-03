// admin_honor_delete.go defines the request and response DTOs for deleting one
// honor definition.

package v1

import "github.com/gogf/gf/v2/frame/g"

// DeleteHonorReq is the request for deleting one honor definition.
type DeleteHonorReq struct {
	g.Meta `path:"/plugins/sicau-niu/admin/honors/{id}" method:"delete" tags:"Sicau Niu Admin" summary:"Delete honor definition" dc:"Soft-delete one honor definition; it is recoverable. Protected by host unified permission check." permission:"sicau-niu:honor:delete"`
	Id     int64 `json:"id" v:"required|min:1" dc:"Honor definition ID from the path" eg:"1"`
}

// DeleteHonorRes is the response for deleting one honor definition. It is
// intentionally empty; success is conveyed by the absence of an error.
type DeleteHonorRes struct{}
