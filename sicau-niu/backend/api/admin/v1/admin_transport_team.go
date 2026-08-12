// This file defines operator-facing cloud-moving team APIs.

package v1

import "github.com/gogf/gf/v2/frame/g"

type ListTransportTeamsReq struct {
	g.Meta   `path:"/plugins/sicau-niu/admin/transport-teams" method:"get" tags:"寻牛运营" summary:"分页查询云搬牛团" dc:"List effective or invalid cloud-moving teams with stored contribution totals. Uses bounded pagination and batch creator projection." permission:"sicau-niu:transport:list"`
	PageNum  int    `json:"pageNum" dc:"Page number; defaults to 1" eg:"1"`
	PageSize int    `json:"pageSize" dc:"Page size; defaults to 20 and is capped at 100" eg:"20"`
	Name     string `json:"name" dc:"Optional fuzzy team-name filter" eg:"川农"`
	Status   string `json:"status" dc:"Optional status filter: effective or invalid" eg:"effective"`
}

type ListTransportTeamsRes struct {
	List  []*AdminTransportTeam `json:"list" dc:"Page of cloud-moving teams" eg:"[]"`
	Total int                   `json:"total" dc:"Matched team count" eg:"1"`
}

type UpdateTransportTeamNameReq struct {
	g.Meta `path:"/plugins/sicau-niu/admin/transport-teams/{id}/name" method:"put" tags:"寻牛运营" summary:"修改云搬牛团名" dc:"Update a team display name through the authenticated operator surface. Effective team names must remain unique; invalid team names may be reused. Renaming does not refresh team activity." permission:"sicau-niu:transport:update" operLog:"update"`
	Id     int64  `json:"id" v:"required|min:1" dc:"Team ID" eg:"12"`
	Name   string `json:"name" v:"required|max-length:64" dc:"New team display name" eg:"川农云搬牛先锋团"`
}

type UpdateTransportTeamNameRes struct{}
