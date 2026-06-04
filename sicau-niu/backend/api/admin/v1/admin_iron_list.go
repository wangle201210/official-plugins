// admin_iron_list.go defines the request and response DTOs for the operator
// iron-cow identifier list query.

package v1

import "github.com/gogf/gf/v2/frame/g"

// ListIronReq is the request for the operator iron-cow list query.
type ListIronReq struct {
	g.Meta   `path:"/plugins/sicau-niu/admin/iron" method:"get" tags:"Sicau Niu Admin" summary:"List iron-cows" dc:"List registered iron-cow identifiers with optional code/name fuzzy filtering, ordered by ID descending, with DB-side pagination. Protected by host unified permission check." permission:"sicau-niu:iron:list"`
	Keyword  string `json:"keyword" dc:"Fuzzy filter by iron-cow code or name; lists all when omitted" eg:"IRON-01"`
	PageNum  int    `json:"pageNum" dc:"Page number; defaults to 1 when omitted or non-positive" eg:"1"`
	PageSize int    `json:"pageSize" dc:"Items per page; defaults to 10 and is capped at 100" eg:"10"`
}

// ListIronRes is the response for the operator iron-cow list query.
type ListIronRes struct {
	List  []*IronItem `json:"list" dc:"Iron-cows on the current page" eg:"[]"`
	Total int         `json:"total" dc:"Total number of matching iron-cows" eg:"1"`
}

// IronItem defines one iron-cow row for the operator console. The real-time
// location fields are written by the C4 bonus flow and are not maintained here.
type IronItem struct {
	Id        int64   `json:"id" dc:"Iron-cow ID" eg:"1"`
	Code      string  `json:"code" dc:"Iron-cow device identifier, unique among active iron-cows" eg:"IRON-01"`
	Name      string  `json:"name" dc:"Iron-cow display name" eg:"图书馆铁牛"`
	LastLat   float64 `json:"lastLat" dc:"Latest GPS latitude written by the C4 bonus flow; 0 when unset" eg:"30.123456"`
	LastLng   float64 `json:"lastLng" dc:"Latest GPS longitude written by the C4 bonus flow; 0 when unset" eg:"103.123456"`
	LocatedAt *int64  `json:"locatedAt" dc:"Latest location pull time as Unix timestamp in milliseconds; null when unset" eg:"1776333600000"`
	Remark    string  `json:"remark" dc:"Iron-cow remark" eg:"门口入口处"`
	CreatedAt *int64  `json:"createdAt" dc:"Creation time as Unix timestamp in milliseconds" eg:"1776333600000"`
	UpdatedAt *int64  `json:"updatedAt" dc:"Update time as Unix timestamp in milliseconds" eg:"1776333900000"`
}
