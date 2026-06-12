// admin_college_update.go defines the request and response DTOs for updating one
// college dictionary entry.

package v1

import "github.com/gogf/gf/v2/frame/g"

// UpdateCollegeReq is the request for updating one college.
type UpdateCollegeReq struct {
	g.Meta `path:"/plugins/sicau-niu/admin/colleges/{id}" method:"put" tags:"Sicau Niu Admin" summary:"修改院系" dc:"Update a college's name and sort order. The name must stay unique among active colleges. Protected by host unified permission check." permission:"sicau-niu:college:update"`
	Id     int64  `json:"id" v:"required|min:1" dc:"College ID from the path" eg:"3"`
	Name   string `json:"name" v:"required|length:1,64" dc:"College name; must be unique among active colleges" eg:"信息工程学院"`
	Sort   int    `json:"sort" dc:"Display sort order, smaller first" eg:"1"`
}

// UpdateCollegeRes is the response for updating one college. It is intentionally
// empty; success is conveyed by the absence of an error.
type UpdateCollegeRes struct{}
