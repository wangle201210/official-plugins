// admin_honor_list.go defines the request and response DTOs for the operator
// honor-definition list query.

package v1

import "github.com/gogf/gf/v2/frame/g"

// ListHonorReq is the request for the operator honor-definition list query.
type ListHonorReq struct {
	g.Meta     `path:"/plugins/sicau-niu/admin/honors" method:"get" tags:"Sicau Niu Admin" summary:"List honor definitions" dc:"List honor definitions with optional code/name fuzzy filtering and honor-type/unlock-type filtering, ordered by sort then ID descending, with DB-side pagination. Protected by host unified permission check." permission:"sicau-niu:honor:list"`
	Keyword    string `json:"keyword" dc:"Fuzzy filter by honor code or name; lists all when omitted" eg:"feed_bronze"`
	HonorType  string `json:"honorType" dc:"Filter by honor type: badge=徽章, avatar_frame=头像框, certificate=证书; lists all when omitted" eg:"badge"`
	UnlockType string `json:"unlockType" dc:"Filter by unlock rule: participation=参与即得, feed_count=喂草次数, activation_count=激活数, category_complete=集齐分类, full_complete=集齐全套; lists all when omitted" eg:"feed_count"`
	PageNum    int    `json:"pageNum" dc:"Page number; defaults to 1 when omitted or non-positive" eg:"1"`
	PageSize   int    `json:"pageSize" dc:"Items per page; defaults to 10 and is capped at 100" eg:"10"`
}

// ListHonorRes is the response for the operator honor-definition list query.
type ListHonorRes struct {
	List  []*HonorItem `json:"list" dc:"Honor definitions on the current page" eg:"[]"`
	Total int          `json:"total" dc:"Total number of matching honor definitions" eg:"1"`
}

// HonorItem defines one honor definition row for the operator console.
type HonorItem struct {
	Id         int64  `json:"id" dc:"Honor definition ID" eg:"1"`
	HonorType  string `json:"honorType" dc:"Honor type: badge=徽章, avatar_frame=头像框, certificate=证书" eg:"badge"`
	Code       string `json:"code" dc:"Honor unique code among active rows" eg:"feed_bronze"`
	Name       string `json:"name" dc:"Honor display name" eg:"青铜喂草师"`
	UnlockType string `json:"unlockType" dc:"Unlock rule: participation=参与即得, feed_count=喂草次数, activation_count=激活数, category_complete=集齐分类, full_complete=集齐全套" eg:"feed_count"`
	Threshold  int    `json:"threshold" dc:"Threshold for count-based unlock rules (feed_count/activation_count); 0 otherwise" eg:"10"`
	Category   string `json:"category" dc:"Card category for the category_complete unlock rule: person, event, research, college, spirit; empty otherwise" eg:"person"`
	ImagePath  string `json:"imagePath" dc:"Honor image/template storage path; empty when none" eg:"/upload/honor/bronze.png"`
	Sort       int    `json:"sort" dc:"Display sort order, smaller first" eg:"1"`
	CreatedAt  *int64 `json:"createdAt" dc:"Creation time as Unix timestamp in milliseconds" eg:"1776333600000"`
	UpdatedAt  *int64 `json:"updatedAt" dc:"Update time as Unix timestamp in milliseconds" eg:"1776333900000"`
}
