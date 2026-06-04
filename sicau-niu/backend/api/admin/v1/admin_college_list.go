// admin_college_list.go defines the request and response DTOs for the operator
// college dictionary list query.

package v1

import "github.com/gogf/gf/v2/frame/g"

// ListCollegesReq is the request for the operator college list query.
type ListCollegesReq struct {
	g.Meta   `path:"/plugins/sicau-niu/admin/colleges" method:"get" tags:"Sicau Niu Admin" summary:"List colleges" dc:"List the college dictionary with optional fuzzy name filtering, ordered by sort then ID, with DB-side pagination. Protected by host unified permission check." permission:"sicau-niu:college:list"`
	Keyword  string `json:"keyword" dc:"Fuzzy filter by college name; lists all colleges when omitted" eg:"信息"`
	PageNum  int    `json:"pageNum" dc:"Page number; defaults to 1 when omitted or non-positive" eg:"1"`
	PageSize int    `json:"pageSize" dc:"Items per page; defaults to 10 and is capped at 100" eg:"10"`
}

// ListCollegesRes is the response for the operator college list query.
type ListCollegesRes struct {
	List  []*CollegeItem `json:"list" dc:"Colleges on the current page" eg:"[]"`
	Total int            `json:"total" dc:"Total number of matching colleges" eg:"1"`
}

// CollegeItem defines one college row for the operator console.
type CollegeItem struct {
	Id        int64  `json:"id" dc:"College ID" eg:"3"`
	Name      string `json:"name" dc:"College name" eg:"信息工程学院"`
	Sort      int    `json:"sort" dc:"Display sort order, smaller first" eg:"1"`
	CreatedAt *int64 `json:"createdAt" dc:"Creation time as Unix timestamp in milliseconds" eg:"1776333600000"`
	UpdatedAt *int64 `json:"updatedAt" dc:"Update time as Unix timestamp in milliseconds" eg:"1776333900000"`
}
