// player_profile.go defines the request and response DTOs for reading and
// updating the current player's own identity profile.

package v1

import "github.com/gogf/gf/v2/frame/g"

// GetProfileReq is the request for reading the current player's profile.
type GetProfileReq struct {
	g.Meta `path:"/plugins/sicau-niu/player/profile" method:"get" tags:"寻牛小程序" summary:"获取当前玩家资料" dc:"Return the current player's own identity profile. Requires a valid player token; only the authenticated player's data is returned."`
}

// GetProfileRes is the response for reading the current player's profile.
type GetProfileRes struct {
	PlayerId       int64  `json:"playerId" dc:"Player ID" eg:"1"`
	Phone          string `json:"phone" dc:"Bound phone number; empty when not yet bound" eg:"13800138000"`
	Nickname       string `json:"nickname" dc:"Player nickname" eg:"川农牛同学"`
	Avatar         string `json:"avatar" dc:"Player avatar URL" eg:"https://example.com/a.png"`
	IdentityType   string `json:"identityType" dc:"Identity tag: student=在校生, alumni=校友, friend=川农好友; empty when not yet set" eg:"student"`
	CollegeId      int64  `json:"collegeId" dc:"Selected college ID; 0 means none" eg:"3"`
	Grade          int    `json:"grade" dc:"Grade number filled by students; 0 means unset" eg:"2024"`
	GraduationYear int    `json:"graduationYear" dc:"Graduation year filled by alumni; 0 means unset" eg:"2018"`
	Level          int    `json:"level" dc:"Player level derived from cumulative effective feeding" eg:"3"`
	Exp            int64  `json:"exp" dc:"Cumulative effective feeding experience" eg:"245"`
	CreatedAt      *int64 `json:"createdAt" dc:"Account creation time as Unix timestamp in milliseconds" eg:"1776333600000"`
	UpdatedAt      *int64 `json:"updatedAt" dc:"Profile update time as Unix timestamp in milliseconds" eg:"1776333900000"`
}

// UpdateProfileReq is the request for updating the current player's profile.
type UpdateProfileReq struct {
	g.Meta         `path:"/plugins/sicau-niu/player/profile" method:"put" tags:"寻牛小程序" summary:"更新当前玩家资料" dc:"Update the current player's nickname, identity tag, college, grade and graduation year. Students must select an existing college and a positive grade; alumni and friends may omit college/grade. Requires a valid player token."`
	Nickname       string `json:"nickname" dc:"Player nickname" eg:"川农牛同学"`
	Avatar         string `json:"avatar" dc:"Player avatar URL" eg:"https://example.com/a.png"`
	IdentityType   string `json:"identityType" v:"required" dc:"Identity tag: student=在校生, alumni=校友, friend=川农好友" eg:"student"`
	CollegeId      int64  `json:"collegeId" dc:"Selected college ID; required for students, optional for alumni/friend; 0 clears the selection for non-students" eg:"3"`
	Grade          int    `json:"grade" dc:"Grade number; required positive for students, ignored for alumni/friend" eg:"2024"`
	GraduationYear int    `json:"graduationYear" dc:"Graduation year for alumni; 0 means unset; validated against a sane range when non-zero" eg:"2018"`
}

// UpdateProfileRes is the response for updating the current player's profile. It
// is intentionally empty; success is conveyed by the absence of an error.
type UpdateProfileRes struct{}
