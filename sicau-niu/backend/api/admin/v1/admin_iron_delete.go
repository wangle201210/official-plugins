// admin_iron_delete.go defines the request and response DTOs for deleting one
// iron-cow registration.

package v1

import "github.com/gogf/gf/v2/frame/g"

// DeleteIronReq is the request for deleting one iron-cow registration.
type DeleteIronReq struct {
	g.Meta `path:"/plugins/sicau-niu/admin/iron/{id}" method:"delete" tags:"Sicau Niu Admin" summary:"删除铁牛" dc:"Soft-delete an iron-cow registration. Recoverable. Protected by host unified permission check." permission:"sicau-niu:iron:delete"`
	Id     int64 `json:"id" v:"required|min:1" dc:"Iron-cow ID from the path" eg:"1"`
}

// DeleteIronRes is the response for deleting one iron-cow. It is intentionally
// empty; success is conveyed by the absence of an error.
type DeleteIronRes struct{}
