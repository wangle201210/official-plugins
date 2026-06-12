// admin_college_create.go defines the request and response DTOs for creating one
// college dictionary entry.

package v1

import "github.com/gogf/gf/v2/frame/g"

// CreateCollegeReq is the request for creating one college.
type CreateCollegeReq struct {
	g.Meta `path:"/plugins/sicau-niu/admin/colleges" method:"post" tags:"Sicau Niu Admin" summary:"新增院系" dc:"Create a college dictionary entry with a unique name. Protected by host unified permission check." permission:"sicau-niu:college:create"`
	Name   string `json:"name" v:"required|length:1,64" dc:"College name; must be unique among active colleges" eg:"信息工程学院"`
	Sort   int    `json:"sort" dc:"Display sort order, smaller first; defaults to 0" eg:"1"`
}

// CreateCollegeRes is the response for creating one college.
type CreateCollegeRes struct {
	Id int64 `json:"id" dc:"The newly created college ID" eg:"3"`
}
