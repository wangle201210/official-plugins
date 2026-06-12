// admin_niu_create.go defines the request and response DTOs for creating one
// cattle, including its online-time visibility configuration.

package v1

import "github.com/gogf/gf/v2/frame/g"

// CreateNiuReq is the request for creating one cattle.
type CreateNiuReq struct {
	g.Meta          `path:"/plugins/sicau-niu/admin/niu" method:"post" tags:"Sicau Niu Admin" summary:"新增牛只" dc:"Create a cattle with a unique code. Common cattle omit name/subtype/college; special cattle carry a name and subtype, and college subtype cattle must link an existing college. Status defaults to inactive. Protected by host unified permission check." permission:"sicau-niu:niu:create"`
	Code            string  `json:"code" v:"required|length:1,64" dc:"Cattle serial code; must be unique among active cattle" eg:"N001"`
	NiuType         string  `json:"niuType" v:"required" dc:"Cattle type: common=普通牛, special=特殊牛" eg:"special"`
	SpecialSubtype  string  `json:"specialSubtype" dc:"Special subtype: college=学院, contribution=贡献, alumni=校友, spirit=精神; required for special cattle, must be empty for common" eg:"college"`
	Name            string  `json:"name" dc:"Cattle name; required for special cattle, optional for common" eg:"信息工程学院牛"`
	CollegeId       int64   `json:"collegeId" dc:"Linked college ID; required for college subtype cattle, 0 otherwise" eg:"3"`
	Lat             float64 `json:"lat" dc:"GPS latitude anchor in GCJ-02" eg:"30.123456"`
	Lng             float64 `json:"lng" dc:"GPS longitude anchor in GCJ-02" eg:"103.123456"`
	OnlineAt        *int64  `json:"onlineAt" dc:"Scheduled online time as Unix timestamp in milliseconds; null means not yet online" eg:"1776333600000"`
	VisibleWeekdays string  `json:"visibleWeekdays" dc:"Optional visible weekdays as comma-separated ISO weekday numbers, e.g. 1,3,5; empty when unset" eg:"1,3,5"`
	VisibleStart    string  `json:"visibleStart" dc:"Optional visible window start as HH:MM; empty when unset" eg:"08:00"`
	VisibleEnd      string  `json:"visibleEnd" dc:"Optional visible window end as HH:MM; empty when unset" eg:"20:00"`
}

// CreateNiuRes is the response for creating one cattle.
type CreateNiuRes struct {
	Id int64 `json:"id" dc:"The newly created cattle ID" eg:"1"`
}
