// admin_niu_list.go defines the request and response DTOs for the operator
// cattle list query, including the batch-assembled college name and main-card
// binding flag.

package v1

import "github.com/gogf/gf/v2/frame/g"

// ListNiuReq is the request for the operator cattle list query.
type ListNiuReq struct {
	g.Meta   `path:"/plugins/sicau-niu/admin/niu" method:"get" tags:"Sicau Niu Admin" summary:"List cattle" dc:"List cattle with optional code/name fuzzy filtering and type filtering, ordered by ID descending, with DB-side pagination. The college name and main-card binding flag are batch-assembled to avoid N+1 queries. Protected by host unified permission check." permission:"sicau-niu:niu:list"`
	Keyword  string `json:"keyword" dc:"Fuzzy filter by cattle code or name; lists all when omitted" eg:"N001"`
	NiuType  string `json:"niuType" dc:"Filter by cattle type: common=普通牛, special=特殊牛; lists all when omitted" eg:"special"`
	PageNum  int    `json:"pageNum" dc:"Page number; defaults to 1 when omitted or non-positive" eg:"1"`
	PageSize int    `json:"pageSize" dc:"Items per page; defaults to 10 and is capped at 100" eg:"10"`
}

// ListNiuRes is the response for the operator cattle list query.
type ListNiuRes struct {
	List  []*NiuItem `json:"list" dc:"Cattle on the current page" eg:"[]"`
	Total int        `json:"total" dc:"Total number of matching cattle" eg:"1"`
}

// NiuItem defines one cattle row for the operator console.
type NiuItem struct {
	Id              int64   `json:"id" dc:"Cattle ID" eg:"1"`
	Code            string  `json:"code" dc:"Cattle serial code, unique among active cattle" eg:"N001"`
	NiuType         string  `json:"niuType" dc:"Cattle type: common=普通牛, special=特殊牛" eg:"special"`
	SpecialSubtype  string  `json:"specialSubtype" dc:"Special subtype: college=学院, contribution=贡献, alumni=校友, spirit=精神; empty for common cattle" eg:"college"`
	Name            string  `json:"name" dc:"Cattle name; used by special cattle, empty for common" eg:"信息工程学院牛"`
	CollegeId       int64   `json:"collegeId" dc:"Linked college ID for college cattle; 0 means none" eg:"3"`
	CollegeName     string  `json:"collegeName" dc:"Linked college name, batch-assembled; empty when unlinked or college missing" eg:"信息工程学院"`
	Lat             float64 `json:"lat" dc:"GPS latitude anchor" eg:"30.123456"`
	Lng             float64 `json:"lng" dc:"GPS longitude anchor" eg:"103.123456"`
	OnlineAt        *int64  `json:"onlineAt" dc:"Scheduled online time as Unix timestamp in milliseconds; null means not yet online" eg:"1776333600000"`
	VisibleWeekdays string  `json:"visibleWeekdays" dc:"Optional visible weekdays as comma-separated ISO weekday numbers, e.g. 1,3,5; empty when unset" eg:"1,3,5"`
	VisibleStart    string  `json:"visibleStart" dc:"Optional visible window start as HH:MM; empty when unset" eg:"08:00"`
	VisibleEnd      string  `json:"visibleEnd" dc:"Optional visible window end as HH:MM; empty when unset" eg:"20:00"`
	Status          string  `json:"status" dc:"Cattle status: inactive=未激活, active=已激活" eg:"inactive"`
	HasCard         bool    `json:"hasCard" dc:"Whether the cattle has a bound main card, batch-assembled" eg:"true"`
	CreatedAt       *int64  `json:"createdAt" dc:"Creation time as Unix timestamp in milliseconds" eg:"1776333600000"`
	UpdatedAt       *int64  `json:"updatedAt" dc:"Update time as Unix timestamp in milliseconds" eg:"1776333900000"`
}
