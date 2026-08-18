// player_rankings.go defines the request and response DTOs for player-facing
// leaderboards. The feeding response contains both player and cattle Top-N views;
// college and SICAU-friend boards remain separate projections. Every board
// aggregates the C4 feeding effect on the database side with a Top-N cap.

package v1

import "github.com/gogf/gf/v2/frame/g"

// FeedRankingReq is the request for player and cattle feeding leaderboards.
type FeedRankingReq struct {
	g.Meta `path:"/plugins/sicau-niu/player/rankings/feed" method:"get" tags:"寻牛小程序" summary:"查询喂草排行榜" dc:"Return player and cattle feeding leaderboards. Players and cattle are ranked by total feeding effect (SUM of effect_amount) descending using competition ranks, so equal totals share a rank and the next rank skips (for example 1, 1, 3). Both lists are capped at the configured Top-N; the player board also includes the requesting player's own rank and total. Aggregation runs on the database side. Requires a valid player token."`
}

// FeedRankingRes is the response for player and cattle feeding leaderboards.
type FeedRankingRes struct {
	List    []*FeedRankItem `json:"list" dc:"Top-N players ordered by total feeding effect descending" eg:"[]"`
	Self    *SelfRank       `json:"self" dc:"The requesting player's own rank and total; rank is 0 when the player has no feeding record" eg:"null"`
	NiuList []*NiuRankItem  `json:"niuList" dc:"Top-N cattle ordered by accumulated feeding effect descending" eg:"[]"`
}

// FeedRankItem defines one player row on the personal feeding leaderboard.
type FeedRankItem struct {
	Rank     int    `json:"rank" dc:"1-based competition rank; equal totals share the same rank and the next rank skips" eg:"1"`
	UserId   int64  `json:"userId" dc:"Player ID" eg:"1"`
	Nickname string `json:"nickname" dc:"Player nickname; empty when unset" eg:"川农牛仔"`
	Total    int64  `json:"total" dc:"Total feeding effect accumulated by the player" eg:"1500"`
}

// NiuRankItem defines one cattle row on the feeding leaderboard.
type NiuRankItem struct {
	Rank    int    `json:"rank" dc:"1-based competition rank; equal totals share the same rank and the next rank skips" eg:"1"`
	NiuId   int64  `json:"niuId" dc:"Cattle ID" eg:"7"`
	NiuCode string `json:"niuCode" dc:"Cattle serial code" eg:"NIU-007"`
	NiuName string `json:"niuName" dc:"Cattle display name; empty when the cattle has no configured name" eg:"信息牛"`
	Total   int64  `json:"total" dc:"Total feeding effect accumulated by the cattle" eg:"3200"`
}

// CollegeRankingReq is the request for the college leaderboard.
type CollegeRankingReq struct {
	g.Meta `path:"/plugins/sicau-niu/player/rankings/college" method:"get" tags:"寻牛小程序" summary:"查询院系排行榜" dc:"Return the college leaderboard: enrolled-student players are aggregated by their college and each college's total feeding effect (SUM of effect_amount) is ranked descending, capped at the configured Top-N. Aggregation runs on the database side. Requires a valid player token."`
}

// CollegeRankingRes is the response for the college leaderboard.
type CollegeRankingRes struct {
	List []*CollegeRankItem `json:"list" dc:"Top-N colleges ordered by total feeding effect descending" eg:"[]"`
}

// CollegeRankItem defines one college row on the college leaderboard.
type CollegeRankItem struct {
	Rank        int    `json:"rank" dc:"1-based rank position on the board" eg:"1"`
	CollegeId   int64  `json:"collegeId" dc:"College ID" eg:"3"`
	CollegeName string `json:"collegeName" dc:"College name; empty when the college is missing" eg:"信息工程学院"`
	Total       int64  `json:"total" dc:"Total feeding effect accumulated by enrolled students of the college" eg:"5200"`
}

// FriendRankingReq is the request for the SICAU-friend leaderboard.
type FriendRankingReq struct {
	g.Meta `path:"/plugins/sicau-niu/player/rankings/friend" method:"get" tags:"寻牛小程序" summary:"查询川农好友榜" dc:"Return the SICAU-friend leaderboard: players whose identity is SICAU-friend (社会好友) ranked by personal total feeding effect (SUM of effect_amount) descending using competition ranks, so equal totals share a rank and the next rank skips. The list is capped at the configured Top-N and includes the requesting player's own rank and total when eligible. Aggregation runs on the database side. Requires a valid player token."`
}

// FriendRankingRes is the response for the SICAU-friend leaderboard.
type FriendRankingRes struct {
	List []*FeedRankItem `json:"list" dc:"Top-N SICAU-friend players ordered by total feeding effect descending" eg:"[]"`
	Self *SelfRank       `json:"self" dc:"The requesting player's own rank and total; rank is 0 when the player is not a SICAU-friend or has no feeding record" eg:"null"`
}

// SelfRank defines the requesting player's own rank and total on a board.
type SelfRank struct {
	Rank  int   `json:"rank" dc:"1-based competition rank; equal totals share a rank; 0 when the player is off the board (no contribution or not eligible)" eg:"7"`
	Total int64 `json:"total" dc:"Total feeding effect accumulated by the player" eg:"120"`
}
