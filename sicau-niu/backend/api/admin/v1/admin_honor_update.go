// admin_honor_update.go defines the request and response DTOs for updating one
// honor definition.

package v1

import "github.com/gogf/gf/v2/frame/g"

// UpdateHonorReq is the request for updating one honor definition.
type UpdateHonorReq struct {
	g.Meta     `path:"/plugins/sicau-niu/admin/honors/{id}" method:"put" tags:"Sicau Niu Admin" summary:"Update honor definition" dc:"Update a honor definition's attributes. The code must stay unique among active honor definitions; honor-type/unlock-type/threshold/category rules match creation. Protected by host unified permission check." permission:"sicau-niu:honor:update"`
	Id         int64  `json:"id" v:"required|min:1" dc:"Honor definition ID from the path" eg:"1"`
	HonorType  string `json:"honorType" v:"required" dc:"Honor type: badge=徽章, avatar_frame=头像框, certificate=证书" eg:"badge"`
	Code       string `json:"code" v:"required|length:1,64" dc:"Honor unique code among active honor definitions" eg:"feed_bronze"`
	Name       string `json:"name" v:"required|length:1,64" dc:"Honor display name" eg:"青铜喂草师"`
	UnlockType string `json:"unlockType" v:"required" dc:"Unlock rule: participation=参与即得, feed_count=喂草次数阈值, activation_count=激活数阈值, category_complete=集齐分类, full_complete=集齐全套" eg:"feed_count"`
	Threshold  int    `json:"threshold" dc:"Threshold for count-based unlock rules (feed_count/activation_count); must be positive for those rules, 0 otherwise" eg:"10"`
	Category   string `json:"category" dc:"Card category for the category_complete unlock rule: person, event, research, college, spirit; required for category_complete, must be empty otherwise" eg:"person"`
	ImagePath  string `json:"imagePath" dc:"Honor image/template storage path; empty when none" eg:"/upload/honor/bronze.png"`
	Sort       int    `json:"sort" dc:"Display sort order, smaller first" eg:"1"`
}

// UpdateHonorRes is the response for updating one honor definition. It is
// intentionally empty; success is conveyed by the absence of an error.
type UpdateHonorRes struct{}
