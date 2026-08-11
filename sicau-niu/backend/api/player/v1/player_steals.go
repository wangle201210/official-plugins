// player_steals.go defines the request and response DTOs for the steal-grass
// feature: the player's daily random stealable-target list (GET) and the steal
// action (POST). The target list is a deterministic per-player, per-day random
// pick of other players; a steal is only allowed against a target in that list
// and is bounded by a daily count limit.

package v1

import "github.com/gogf/gf/v2/frame/g"

// StealTargetsReq is the request for the player's daily stealable-target list.
type StealTargetsReq struct {
	g.Meta `path:"/plugins/sicau-niu/player/steal-targets" method:"get" tags:"寻牛小程序" summary:"查询今日可偷草目标" dc:"Return the authenticated player's deterministic daily random list of stealable players. Each item includes display fields and whether the current player already stole that target during the Beijing natural day. Requires a valid player token."`
}

// StealTargetsRes is the response for the player's daily stealable-target list.
type StealTargetsRes struct {
	List []*StealTargetItem `json:"list" dc:"Today's stealable targets" eg:"[]"`
}

// StealTargetItem defines one stealable target projected for the player.
type StealTargetItem struct {
	UserId      int64  `json:"userId" dc:"Stealable target player ID" eg:"2"`
	Nickname    string `json:"nickname" dc:"Target player nickname" eg:"川农同学"`
	Avatar      string `json:"avatar" dc:"Target player avatar URL" eg:"https://example.com/avatar.png"`
	StolenToday bool   `json:"stolenToday" dc:"Whether the current player already stole this target today" eg:"false"`
}

// StealReq is the request for stealing grass from a target.
type StealReq struct {
	g.Meta       `path:"/plugins/sicau-niu/player/steals" method:"post" tags:"寻牛小程序" summary:"偷取目标草料" dc:"Steal a small random amount of grass from a target that appears in the player's daily stealable list. The action is rejected when the target is not in today's list or the player has reached the daily steal limit. The stolen amount is debited from the target and credited to the player in one transaction (both ledger transactions), and the target receives a stolen-notification message. Requires a valid player token."`
	TargetUserId int64  `json:"targetUserId" v:"required" dc:"Target player ID to steal from; must be in today's stealable list" eg:"2"`
	RequestId    string `json:"requestId" v:"max-length:64" dc:"Optional client idempotency key (max 64 chars); resend the same key on a network retry and the server rejects the duplicate instead of stealing twice" eg:"steal-1b2a3c4d"`
}

// StealRes is the response for a successful steal.
type StealRes struct {
	Amount  int   `json:"amount" dc:"Grass amount stolen and credited to the player" eg:"12"`
	Balance int64 `json:"balance" dc:"Player grass balance after the steal" eg:"145"`
}
