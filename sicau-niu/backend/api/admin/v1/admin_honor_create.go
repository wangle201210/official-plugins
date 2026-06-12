// admin_honor_create.go defines the request and response DTOs for creating one
// honor definition.

package v1

import "github.com/gogf/gf/v2/frame/g"

// CreateHonorReq is the request for creating one honor definition.
type CreateHonorReq struct {
	g.Meta     `path:"/plugins/sicau-niu/admin/honors" method:"post" tags:"Sicau Niu Admin" summary:"新增荣誉定义" dc:"Create a honor definition with a unique code. The honor type and unlock rule must be valid enum values; count-based unlock rules (feed_count/activation_count) use threshold; the category_complete rule uses category. Protected by host unified permission check." permission:"sicau-niu:honor:create"`
	HonorType  string `json:"honorType" v:"required" dc:"Honor type: badge=徽章, avatar_frame=头像框, certificate=证书" eg:"badge"`
	Code       string `json:"code" v:"required|length:1,64" dc:"Honor unique code among active honor definitions" eg:"feed_bronze"`
	Name       string `json:"name" v:"required|length:1,64" dc:"Honor display name" eg:"青铜喂草师"`
	UnlockType string `json:"unlockType" v:"required" dc:"Unlock rule: participation=参与即得, feed_count=喂草次数阈值, activation_count=激活数阈值, category_complete=集齐分类, full_complete=集齐全套" eg:"feed_count"`
	Threshold  int    `json:"threshold" dc:"Threshold for count-based unlock rules (feed_count/activation_count); must be positive for those rules, 0 otherwise" eg:"10"`
	Category   string `json:"category" dc:"Card category for the category_complete unlock rule: person, event, research, college, spirit; required for category_complete, must be empty otherwise" eg:"person"`
	ImagePath  string `json:"imagePath" dc:"Honor image/template storage path; empty when none" eg:"/upload/honor/bronze.png"`
	Sort       int    `json:"sort" dc:"Display sort order, smaller first; defaults to 0" eg:"1"`
}

// CreateHonorRes is the response for creating one honor definition.
type CreateHonorRes struct {
	Id int64 `json:"id" dc:"The newly created honor definition ID" eg:"1"`
}
