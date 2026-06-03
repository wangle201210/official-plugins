// player_poster.go defines the request and response DTOs for the activation
// poster composition data: nickname, identity type, cattle code, arrival order, a
// random school-history quote and the campus anniversary badge for an activated
// cattle. The raw identity type is returned for the frontend to map to a label.

package v1

import "github.com/gogf/gf/v2/frame/g"

// PosterReq is the request for the activation poster composition data.
type PosterReq struct {
	g.Meta `path:"/plugins/sicau-niu/player/poster" method:"get" tags:"Sicau Niu Player" summary:"Get activation poster data" dc:"Return the activation poster composition data for a cattle the authenticated player has activated: nickname, raw identity type (frontend maps to a label), cattle code, arrival order, a random enabled school-history quote and the campus anniversary badge. The player must have already activated the cattle, otherwise a business error is returned. Requires a valid player token."`
	NiuId  int64 `json:"niuId" v:"required" dc:"Target cattle ID the player has activated" eg:"1"`
}

// PosterRes is the response for the activation poster composition data.
type PosterRes struct {
	Nickname     string `json:"nickname" dc:"Player nickname" eg:"川农牛同学"`
	IdentityType string `json:"identityType" dc:"Raw identity type for the frontend to map to a label: student=在校生, alumni=校友, friend=川农好友; empty when unset" eg:"student"`
	NiuCode      string `json:"niuCode" dc:"Activated cattle serial code" eg:"NIU-001"`
	OrderNo      int    `json:"orderNo" dc:"The player's arrival order for this cattle, starting at 1" eg:"1"`
	Quote        string `json:"quote" dc:"A random enabled school-history quote; empty when no enabled quote exists" eg:"任重道远，砥砺前行"`
	CampusBadge  string `json:"campusBadge" dc:"Campus anniversary badge text from plugin config; empty when unset" eg:"川农120周年校庆"`
}
