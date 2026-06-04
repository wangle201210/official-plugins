// admin_college_delete.go defines the request and response DTOs for deleting one
// college dictionary entry.

package v1

import "github.com/gogf/gf/v2/frame/g"

// DeleteCollegeReq is the request for deleting one college.
type DeleteCollegeReq struct {
	g.Meta `path:"/plugins/sicau-niu/admin/colleges/{id}" method:"delete" tags:"Sicau Niu Admin" summary:"Delete college" dc:"Soft-delete a college. The deletion is rejected when the college is still selected by any player, to avoid dangling references. Protected by host unified permission check." permission:"sicau-niu:college:delete"`
	Id     int64 `json:"id" v:"required|min:1" dc:"College ID from the path" eg:"3"`
}

// DeleteCollegeRes is the response for deleting one college. It is intentionally
// empty; success is conveyed by the absence of an error.
type DeleteCollegeRes struct{}
