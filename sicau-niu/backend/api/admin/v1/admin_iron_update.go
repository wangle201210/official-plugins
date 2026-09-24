// admin_iron_update.go defines the request and response DTOs for updating one
// iron-cow identifier registration.

package v1

import "github.com/gogf/gf/v2/frame/g"

// UpdateIronReq is the request for updating one iron-cow registration.
type UpdateIronReq struct {
	g.Meta       `path:"/plugins/sicau-niu/admin/iron/{id}" method:"put" tags:"Sicau Niu Admin" summary:"修改铁牛" dc:"Update an iron-cow's code, name, remark and proximity-bonus switch. The code must stay unique among active iron-cows. Real-time location fields are not changed here. Protected by host unified permission check." permission:"sicau-niu:iron:update"`
	Id           int64  `json:"id" v:"required|min:1" dc:"Iron-cow ID from the path" eg:"1"`
	Code         string `json:"code" v:"required|length:1,64" dc:"Iron-cow device identifier; must be unique among active iron-cows" eg:"IRON-01"`
	Name         string `json:"name" v:"required|length:1,128" dc:"Iron-cow display name" eg:"图书馆铁牛"`
	BonusEnabled *bool  `json:"bonusEnabled" dc:"Whether this iron cow grants the proximity bonus; omitted keeps its current setting, false disables it immediately" eg:"true"`
	Remark       string `json:"remark" dc:"Iron-cow remark" eg:"门口入口处"`
}

// UpdateIronRes is the response for updating one iron-cow. It is intentionally
// empty; success is conveyed by the absence of an error.
type UpdateIronRes struct{}
