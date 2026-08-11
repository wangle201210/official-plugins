// player_checkin.go defines the request and response DTOs for the daily
// check-in action: the authenticated player checks in once per natural day to
// receive a random grass grant credited to their ledger account. The response
// carries the granted amount and the resulting balance.

package v1

import "github.com/gogf/gf/v2/frame/g"

// CheckinReq is the request for the player's daily check-in.
type CheckinReq struct {
	g.Meta    `path:"/plugins/sicau-niu/player/checkin" method:"post" tags:"寻牛小程序" summary:"每日签到领取草料" dc:"Check in once for the current Beijing natural day to receive a random grass grant credited to the authenticated player's ledger account. Reusing requestId returns the first successful amount and balance without a second credit. Requires a valid player token."`
	RequestId string `json:"requestId" v:"required|max-length:64" dc:"Player-scoped idempotency key" eg:"checkin-1b2a3c4d"`
}

// CheckinRes is the response for a successful daily check-in.
type CheckinRes struct {
	Amount  int   `json:"amount" dc:"Grass amount granted by this check-in" eg:"33"`
	Balance int64 `json:"balance" dc:"Player grass balance after the check-in" eg:"133"`
}
