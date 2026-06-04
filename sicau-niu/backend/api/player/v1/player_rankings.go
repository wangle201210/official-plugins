// player_rankings.go defines the request and response DTOs for the three player
// leaderboards: the personal feeding board (Top-N + self rank), the college board
// (Top-N) and the SICAU-friend board (Top-N + self rank). All boards aggregate
// the C4 feeding effect on the database side with a Top-N cap; the personal and
// friend boards additionally return the requesting player's own rank.

package v1

import "github.com/gogf/gf/v2/frame/g"

// FeedRankingReq is the request for the personal feeding leaderboard.
type FeedRankingReq struct {
	g.Meta `path:"/plugins/sicau-niu/player/rankings/feed" method:"get" tags:"Sicau Niu Player" summary:"Personal feeding leaderboard" dc:"Return the personal feeding leaderboard: players ranked by their total feeding effect (SUM of effect_amount) descending, capped at the configured Top-N, plus the requesting player's own rank and total. Aggregation runs on the database side. Requires a valid player token."`
}

// FeedRankingRes is the response for the personal feeding leaderboard.
type FeedRankingRes struct {
	List []*FeedRankItem `json:"list" dc:"Top-N players ordered by total feeding effect descending" eg:"[]"`
	Self *SelfRank       `json:"self" dc:"The requesting player's own rank and total; rank is 0 when the player has no feeding record" eg:"null"`
}

// FeedRankItem defines one player row on the personal feeding leaderboard.
type FeedRankItem struct {
	Rank     int    `json:"rank" dc:"1-based rank position on the board" eg:"1"`
	UserId   int64  `json:"userId" dc:"Player ID" eg:"1"`
	Nickname string `json:"nickname" dc:"Player nickname; empty when unset" eg:"川农牛仔"`
	Total    int64  `json:"total" dc:"Total feeding effect accumulated by the player" eg:"1500"`
}

// CollegeRankingReq is the request for the college leaderboard.
type CollegeRankingReq struct {
	g.Meta `path:"/plugins/sicau-niu/player/rankings/college" method:"get" tags:"Sicau Niu Player" summary:"College leaderboard" dc:"Return the college leaderboard: enrolled-student players are aggregated by their college and each college's total feeding effect (SUM of effect_amount) is ranked descending, capped at the configured Top-N. Aggregation runs on the database side. Requires a valid player token."`
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
	g.Meta `path:"/plugins/sicau-niu/player/rankings/friend" method:"get" tags:"Sicau Niu Player" summary:"SICAU-friend leaderboard" dc:"Return the SICAU-friend leaderboard: players whose identity is SICAU-friend (社会好友) ranked by their personal total feeding effect (SUM of effect_amount) descending, capped at the configured Top-N, plus the requesting player's own rank and total when the player is a SICAU-friend. Aggregation runs on the database side. Requires a valid player token."`
}

// FriendRankingRes is the response for the SICAU-friend leaderboard.
type FriendRankingRes struct {
	List []*FeedRankItem `json:"list" dc:"Top-N SICAU-friend players ordered by total feeding effect descending" eg:"[]"`
	Self *SelfRank       `json:"self" dc:"The requesting player's own rank and total; rank is 0 when the player is not a SICAU-friend or has no feeding record" eg:"null"`
}

// SelfRank defines the requesting player's own rank and total on a board.
type SelfRank struct {
	Rank  int   `json:"rank" dc:"1-based rank of the player; 0 when the player is off the board (no contribution or not eligible)" eg:"7"`
	Total int64 `json:"total" dc:"Total feeding effect accumulated by the player" eg:"120"`
}
