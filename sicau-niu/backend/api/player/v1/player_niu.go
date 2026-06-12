// player_niu.go defines the request and response DTOs for the player-facing
// visible-cattle map list: the currently visible cattle (filtered by online time
// plus optional visible windows) with each cattle's GPS anchor, shared-pool activation state and
// whether the current player has already activated it.

package v1

import "github.com/gogf/gf/v2/frame/g"

// VisibleNiuReq is the request for the current player's visible-cattle map list.
type VisibleNiuReq struct {
	g.Meta `path:"/plugins/sicau-niu/player/niu" method:"get" tags:"寻牛小程序" summary:"查询地图可见牛只" dc:"Return the cattle currently visible to the authenticated player, filtered by online time reached plus optional weekday/time window. Each item carries its GPS anchor, shared-pool activation status and whether the current player has already activated it. The visible set is bounded (≤120) and returned in one response; statuses are batch-assembled to avoid N+1. Requires a valid player token."`
}

// VisibleNiuRes is the response for the current player's visible-cattle map list.
type VisibleNiuRes struct {
	List []*VisibleNiuItem `json:"list" dc:"Currently visible cattle ordered by ID ascending" eg:"[]"`
}

// VisibleNiuItem defines one visible cattle projected for the player map.
type VisibleNiuItem struct {
	Id            int64   `json:"id" dc:"Cattle ID" eg:"1"`
	Code          string  `json:"code" dc:"Cattle serial code" eg:"NIU-001"`
	NiuType       string  `json:"niuType" dc:"Cattle type: common=普通牛, special=特殊牛" eg:"special"`
	Name          string  `json:"name" dc:"Cattle name; empty for common cattle" eg:"川农魂"`
	Lat           float64 `json:"lat" dc:"GPS latitude anchor in GCJ-02" eg:"30.123456"`
	Lng           float64 `json:"lng" dc:"GPS longitude anchor in GCJ-02" eg:"103.123456"`
	Status        string  `json:"status" dc:"Shared-pool activation status: inactive=未激活, active=已激活(全员可见)" eg:"active"`
	ActivatedByMe bool    `json:"activatedByMe" dc:"Whether the current player has already activated this cattle" eg:"false"`
}
