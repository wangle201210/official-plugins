// player_checkin.go defines the request and response DTOs for the daily
// check-in action: the authenticated player checks in once per natural day to
// receive a random grass grant credited to their ledger account. The response
// carries the granted amount and the resulting balance.

package v1

import "github.com/gogf/gf/v2/frame/g"

// CheckinReq is the request for the player's daily check-in.
type CheckinReq struct {
	g.Meta `path:"/plugins/sicau-niu/player/checkin" method:"post" tags:"Sicau Niu Player" summary:"Daily check-in to receive grass" dc:"Check in once for the current natural day to receive a random grass grant (default 20–50) credited to the authenticated player's ledger account. A second check-in on the same day is rejected. Requires a valid player token."`
}

// CheckinRes is the response for a successful daily check-in.
type CheckinRes struct {
	Amount  int   `json:"amount" dc:"Grass amount granted by this check-in" eg:"33"`
	Balance int64 `json:"balance" dc:"Player grass balance after the check-in" eg:"133"`
}
