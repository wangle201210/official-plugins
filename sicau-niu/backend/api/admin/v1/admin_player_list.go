// admin_player_list.go defines the request and response DTOs for the operator
// read-only player query.

package v1

import "github.com/gogf/gf/v2/frame/g"

// ListPlayersReq is the request for the operator player list query.
type ListPlayersReq struct {
	g.Meta       `path:"/plugins/sicau-niu/admin/players" method:"get" tags:"Sicau Niu Admin" summary:"查询玩家列表" dc:"Read-only paged query of player basic information with optional nickname and identity-type filtering, ordered by newest first, with DB-side pagination. Protected by host unified permission check." permission:"sicau-niu:player:list"`
	Keyword      string `json:"keyword" dc:"Fuzzy filter by player nickname; lists all players when omitted" eg:"川农"`
	IdentityType string `json:"identityType" dc:"Filter by identity tag: student=在校生, alumni=校友, friend=川农好友; lists all when omitted" eg:"student"`
	PageNum      int    `json:"pageNum" dc:"Page number; defaults to 1 when omitted or non-positive" eg:"1"`
	PageSize     int    `json:"pageSize" dc:"Items per page; defaults to 10 and is capped at 100" eg:"10"`
}

// ListPlayersRes is the response for the operator player list query.
type ListPlayersRes struct {
	List  []*PlayerItem `json:"list" dc:"Players on the current page" eg:"[]"`
	Total int           `json:"total" dc:"Total number of matching players" eg:"1"`
}

// PlayerItem defines one player row for the operator console.
type PlayerItem struct {
	Id             int64  `json:"id" dc:"Player ID" eg:"1"`
	Nickname       string `json:"nickname" dc:"Player nickname" eg:"川农牛同学"`
	Phone          string `json:"phone" dc:"Bound phone number; empty when not bound" eg:"13800138000"`
	IdentityType   string `json:"identityType" dc:"Identity tag: student=在校生, alumni=校友, friend=川农好友; empty when not yet set" eg:"student"`
	CollegeId      int64  `json:"collegeId" dc:"Selected college ID; 0 means none" eg:"3"`
	Grade          int    `json:"grade" dc:"Grade number; 0 means unset" eg:"2024"`
	GraduationYear int    `json:"graduationYear" dc:"Graduation year; 0 means unset" eg:"2018"`
	CreatedAt      *int64 `json:"createdAt" dc:"Account creation time as Unix timestamp in milliseconds" eg:"1776333600000"`
	UpdatedAt      *int64 `json:"updatedAt" dc:"Profile update time as Unix timestamp in milliseconds" eg:"1776333900000"`
}
