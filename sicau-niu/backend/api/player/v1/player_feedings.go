// player_feedings.go defines the request and response DTOs for feeding: the
// player feeds grass to an already-activated cattle (POST), receiving the cattle
// info, a random school-history quote and the bonus breakdown; and the player's
// recent feeding trail (GET, latest 10). Feeding deducts the base amount from the
// player's grass account; an iron-cow proximity bonus may raise the effect.

package v1

import "github.com/gogf/gf/v2/frame/g"

// FeedReq is the request for feeding grass to an activated cattle.
type FeedReq struct {
	g.Meta     `path:"/plugins/sicau-niu/player/feedings" method:"post" tags:"Sicau Niu Player" summary:"Feed grass to an activated cattle" dc:"Feed baseAmount grass to an already-activated cattle. The server checks the player has enough grass, deducts the base amount (ledger feed transaction), and computes the effect: when the cattle anchor is within the iron-bonus distance of any iron cow's current location the coefficient is 1.5 (coefficientBasis=150), otherwise 1.0 (coefficientBasis=100). The response carries the cattle info, a random school-history quote and the bonus breakdown. Feeding a not-yet-activated cattle or feeding beyond the balance is rejected. Requires a valid player token."`
	NiuId      int64 `json:"niuId" v:"required" dc:"Target activated cattle ID to feed" eg:"1"`
	BaseAmount int   `json:"baseAmount" v:"required" dc:"Grass amount to feed (deducted from balance); must be positive and not exceed the balance" eg:"10"`
}

// FeedRes is the response for a successful feeding.
type FeedRes struct {
	NiuId            int64  `json:"niuId" dc:"Fed cattle ID" eg:"1"`
	NiuCode          string `json:"niuCode" dc:"Fed cattle serial code" eg:"NIU-001"`
	NiuName          string `json:"niuName" dc:"Fed cattle name; empty for common cattle" eg:"川农魂"`
	Quote            string `json:"quote" dc:"A random enabled school-history quote; empty when no quote exists" eg:"爱国敬业、艰苦奋斗、团结拼搏、求实创新"`
	BaseAmount       int    `json:"baseAmount" dc:"Original grass amount fed" eg:"10"`
	CoefficientBasis int    `json:"coefficientBasis" dc:"Bonus coefficient in basis of 100: 100=x1.0, 150=x1.5" eg:"150"`
	EffectAmount     int    `json:"effectAmount" dc:"Actual feeding effect = baseAmount * coefficientBasis / 100" eg:"15"`
	IsIronBonus      bool   `json:"isIronBonus" dc:"Whether an iron-cow proximity bonus applied" eg:"true"`
	Balance          int64  `json:"balance" dc:"Player grass balance after the feeding deduction" eg:"123"`
}

// FeedingTrailReq is the request for the player's recent feeding trail.
type FeedingTrailReq struct {
	g.Meta `path:"/plugins/sicau-niu/player/feedings" method:"get" tags:"Sicau Niu Player" summary:"List recent feeding trail" dc:"Return the authenticated player's most recent feeding records (latest 10, newest first), each with the fed cattle code/name, effect amount and feed time. The trail is isolated to the current player and empty when the player has not fed any cattle. Requires a valid player token."`
}

// FeedingTrailRes is the response for the player's recent feeding trail.
type FeedingTrailRes struct {
	List []*FeedingTrailItem `json:"list" dc:"Recent feedings ordered by feed time descending (latest 10)" eg:"[]"`
}

// FeedingTrailItem defines one feeding record projected for the player trail.
type FeedingTrailItem struct {
	NiuId        int64  `json:"niuId" dc:"Fed cattle ID" eg:"1"`
	NiuCode      string `json:"niuCode" dc:"Fed cattle serial code" eg:"NIU-001"`
	NiuName      string `json:"niuName" dc:"Fed cattle name; empty for common cattle" eg:"川农魂"`
	EffectAmount int    `json:"effectAmount" dc:"Actual feeding effect for this record" eg:"15"`
	IsIronBonus  bool   `json:"isIronBonus" dc:"Whether an iron-cow proximity bonus applied to this record" eg:"true"`
	FedAt        *int64 `json:"fedAt" dc:"Feed time as Unix timestamp in milliseconds" eg:"1776333600000"`
}
