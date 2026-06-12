// admin_iron_create.go defines the request and response DTOs for registering one
// iron-cow identifier.

package v1

import "github.com/gogf/gf/v2/frame/g"

// CreateIronReq is the request for registering one iron-cow identifier.
type CreateIronReq struct {
	g.Meta `path:"/plugins/sicau-niu/admin/iron" method:"post" tags:"Sicau Niu Admin" summary:"新增铁牛" dc:"Register an iron-cow with a unique device code. The real-time location is not set here; it is written later by the C4 bonus flow. Protected by host unified permission check." permission:"sicau-niu:iron:create"`
	Code   string `json:"code" v:"required|length:1,64" dc:"Iron-cow device identifier; must be unique among active iron-cows" eg:"IRON-01"`
	Name   string `json:"name" v:"required|length:1,128" dc:"Iron-cow display name" eg:"图书馆铁牛"`
	Remark string `json:"remark" dc:"Iron-cow remark" eg:"门口入口处"`
}

// CreateIronRes is the response for registering one iron-cow.
type CreateIronRes struct {
	Id int64 `json:"id" dc:"The newly created iron-cow ID" eg:"1"`
}
