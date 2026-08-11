// admin_niu_import.go defines the bounded cattle bulk-import API.
package v1

import "github.com/gogf/gf/v2/frame/g"

type ImportNiuReq struct {
	g.Meta    `path:"/plugins/sicau-niu/admin/niu/import" method:"post" tags:"Sicau Niu Admin" summary:"批量导入牛只" dc:"Validate and import at most 200 cattle in one transaction with set-based code and college lookups. Existing codes are updated only when overwrite is true. Protected by host unified permission check." permission:"sicau-niu:niu:create"`
	Items     []*ImportNiuItem `json:"items" v:"required|length:1,200" dc:"Cattle rows to import" eg:"[]"`
	Overwrite bool             `json:"overwrite" dc:"Update existing active cattle matched by code" eg:"false"`
}

type ImportNiuRes struct {
	Created int `json:"created" dc:"Number of newly created cattle" eg:"10"`
	Updated int `json:"updated" dc:"Number of overwritten cattle" eg:"2"`
	Skipped int `json:"skipped" dc:"Number of existing codes skipped" eg:"1"`
}

type ImportNiuItem struct {
	Code            string  `json:"code" v:"required|length:1,64" dc:"Unique cattle code" eg:"N001"`
	NiuType         string  `json:"niuType" v:"required|in:common,special" dc:"Cattle type" eg:"special"`
	SpecialSubtype  string  `json:"specialSubtype" dc:"Special subtype" eg:"college"`
	Name            string  `json:"name" dc:"Cattle display name" eg:"信息工程学院牛"`
	CollegeId       int64   `json:"collegeId" dc:"Linked college ID for college subtype" eg:"3"`
	Lat             float64 `json:"lat" v:"between:-90,90" dc:"GCJ-02 latitude" eg:"30.123456"`
	Lng             float64 `json:"lng" v:"between:-180,180" dc:"GCJ-02 longitude" eg:"103.123456"`
	OnlineAt        *int64  `json:"onlineAt" dc:"Online time as Unix milliseconds" eg:"1776333600000"`
	VisibleWeekdays string  `json:"visibleWeekdays" dc:"Optional ISO weekdays" eg:"1,3,5"`
	VisibleStart    string  `json:"visibleStart" dc:"Optional visibility start HH:MM" eg:"08:00"`
	VisibleEnd      string  `json:"visibleEnd" dc:"Optional visibility end HH:MM" eg:"20:00"`
}
