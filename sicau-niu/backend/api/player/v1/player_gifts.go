// player_gifts.go defines the request and response DTOs for the gift-grass
// action: the player gives some of their own grass to another player. Gifting is
// bounded by a daily count limit and a per-gift minimum amount; the amount is
// debited from the giver and credited to the recipient in one transaction, and
// the recipient receives a gift-notification message.

package v1

import "github.com/gogf/gf/v2/frame/g"

// GiftReq is the request for gifting grass to another player.
type GiftReq struct {
	g.Meta    `path:"/plugins/sicau-niu/player/gifts" method:"post" tags:"寻牛小程序" summary:"赠送草料" dc:"Give amount grass to another player. A required player-scoped requestId makes network retries stable: reusing a successful key returns its first giver balance without another transfer or notification. Each player may gift at most a configured number of times per natural day (default 12) and each gift must be at least the configured minimum (default 12). The action is rejected when the daily limit is reached, the amount is below the minimum, or the balance is insufficient. The amount is debited from the giver and credited to the recipient in one transaction (both ledger transactions), and the recipient receives a gift-received message. Requires a valid player token."`
	ToUserId  int64  `json:"toUserId" v:"required" dc:"Recipient player ID" eg:"2"`
	Amount    int    `json:"amount" v:"required" dc:"Grass amount to gift; must be at least the configured per-gift minimum and not exceed the balance" eg:"12"`
	RequestId string `json:"requestId" v:"required|max-length:64" dc:"Required player-scoped idempotency key (max 64 chars); resend the same key while a network result is unknown and the server returns the first successful giver balance without gifting twice" eg:"gift-1b2a3c4d"`
}

// GiftRes is the response for a successful gift.
type GiftRes struct {
	Balance int64 `json:"balance" dc:"Giver grass balance after the gift" eg:"121"`
}
