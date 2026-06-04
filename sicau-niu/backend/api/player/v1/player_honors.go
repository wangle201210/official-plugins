// player_honors.go defines the request and response DTOs for the player honor
// list. Each honor definition is returned with its read-only unlock status,
// computed from the player's feeding count, activation count and card-collection
// progress. This endpoint never persists a grant (persistent grant is C7).

package v1

import "github.com/gogf/gf/v2/frame/g"

// PlayerHonorsReq is the request for the player honor list.
type PlayerHonorsReq struct {
	g.Meta `path:"/plugins/sicau-niu/player/honors" method:"get" tags:"Sicau Niu Player" summary:"List player honors with unlock status" dc:"Return every honor definition with the requesting player's read-only unlock status, computed from their feeding count, activation count and card-collection completion against each honor's unlock rule. The endpoint is read-only and does not persist any grant. Counts are batch-aggregated to avoid per-honor queries. Requires a valid player token."`
}

// PlayerHonorsRes is the response for the player honor list.
type PlayerHonorsRes struct {
	List []*PlayerHonorItem `json:"list" dc:"Honor definitions with the player's unlock status, ordered by sort then ID" eg:"[]"`
}

// PlayerHonorItem defines one honor definition with the player's unlock status.
type PlayerHonorItem struct {
	Id         int64  `json:"id" dc:"Honor definition ID" eg:"1"`
	HonorType  string `json:"honorType" dc:"Honor type: badge=徽章, avatar_frame=头像框, certificate=证书" eg:"badge"`
	Code       string `json:"code" dc:"Honor unique code" eg:"feed_bronze"`
	Name       string `json:"name" dc:"Honor display name" eg:"青铜喂草师"`
	UnlockType string `json:"unlockType" dc:"Unlock rule: participation=参与即得, feed_count=喂草次数, activation_count=激活数, category_complete=集齐分类, full_complete=集齐全套" eg:"feed_count"`
	Threshold  int    `json:"threshold" dc:"Threshold for count-based unlock rules; 0 otherwise" eg:"10"`
	Category   string `json:"category" dc:"Card category for the category_complete unlock rule; empty otherwise" eg:"person"`
	ImagePath  string `json:"imagePath" dc:"Honor image/template storage path; empty when none" eg:"/upload/honor/bronze.png"`
	Unlocked   bool   `json:"unlocked" dc:"Whether the player has met this honor's unlock rule (read-only computation)" eg:"true"`
}
