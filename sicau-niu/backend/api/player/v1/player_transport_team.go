// This file defines player-facing cloud-moving team lifecycle APIs.

package v1

import "github.com/gogf/gf/v2/frame/g"

type CreateTransportTeamReq struct {
	g.Meta    `path:"/plugins/sicau-niu/player/iron-transport/teams" method:"post" tags:"寻牛小程序" summary:"创建云搬牛团" dc:"Create one effective cloud-moving team and atomically join its creator. Effective team names are unique and become reusable after invalidation. At most 120 effective teams may exist and one player may belong to only one effective team. Reusing requestId does not duplicate the write and returns current state. Requires a valid player token."`
	Name      string `json:"name" v:"required|max-length:64" dc:"Unique name among effective teams; this is the creator's only player-side naming opportunity" eg:"川农云搬牛一团"`
	RequestId string `json:"requestId" v:"required|max-length:64" dc:"Player-scoped idempotency key" eg:"transport-create-1b2a3c4d"`
}

type CreateTransportTeamRes struct {
	*IronTransportState `json:",inline" dc:"State after creating the team" eg:"{}"`
}

type JoinTransportTeamReq struct {
	g.Meta    `path:"/plugins/sicau-niu/player/iron-transport/teams/{id}/join" method:"post" tags:"寻牛小程序" summary:"加入云搬牛团" dc:"Join one effective team without a distance or member-count restriction. A player may belong to only one effective team. Reusing requestId does not duplicate the write and returns current state. Requires a valid player token."`
	Id        int64  `json:"id" v:"required|min:1" dc:"Target team ID" eg:"12"`
	RequestId string `json:"requestId" v:"required|max-length:64" dc:"Player-scoped idempotency key" eg:"transport-join-1b2a3c4d"`
}

type JoinTransportTeamRes struct {
	*IronTransportState `json:",inline" dc:"State after joining the team" eg:"{}"`
}

type GetTransportTeamReq struct {
	g.Meta `path:"/plugins/sicau-niu/player/iron-transport/teams/{id}" method:"get" tags:"寻牛小程序" summary:"查询云搬牛团详情" dc:"Return one effective team summary. Invalid teams are treated as not found. Requires a valid player token."`
	Id     int64 `json:"id" v:"required|min:1" dc:"Team ID" eg:"12"`
}

type GetTransportTeamRes struct {
	Team *IronTransportTeam `json:"team" dc:"Effective team summary" eg:"{}"`
}

type ListTransportMembersReq struct {
	g.Meta   `path:"/plugins/sicau-niu/player/iron-transport/teams/{id}/members" method:"get" tags:"寻牛小程序" summary:"分页查询云搬牛团成员" dc:"Return public member identity and contribution totals only when the current player belongs to the effective team. Raw position reports are not exposed. Requires a valid player token."`
	Id       int64 `json:"id" v:"required|min:1" dc:"Team ID" eg:"12"`
	PageNum  int   `json:"pageNum" dc:"Page number; defaults to 1" eg:"1"`
	PageSize int   `json:"pageSize" dc:"Page size; defaults to 20 and is capped at 100" eg:"20"`
}

type ListTransportMembersRes struct {
	List  []*IronTransportMember `json:"list" dc:"Page of active team members" eg:"[]"`
	Total int                    `json:"total" dc:"Active member count" eg:"1"`
}
