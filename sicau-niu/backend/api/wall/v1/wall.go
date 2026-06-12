// wall.go defines the request and response DTOs for the sicau-niu C6 public
// memorial wall: the first-activator wall, the campus-history highlights and the
// public activity stats. All three endpoints are public (no authentication, no
// player token) and expose only nicknames and activity information; they never
// return phone numbers, openids or device fingerprints. Time fields are returned
// as Unix milliseconds.

package v1

import "github.com/gogf/gf/v2/frame/g"

// FirstActivatorsReq is the request for the public first-activator wall.
type FirstActivatorsReq struct {
	g.Meta `path:"/plugins/sicau-niu/wall/first-activators" method:"get" tags:"Sicau Niu Wall" summary:"公共首发激活墙" dc:"Return the first-activator memorial wall: the players who were the first to activate a cattle (is_first), ordered by activation time ascending and capped at 120. This endpoint is public and requires no authentication. The response exposes only nicknames and activity information; it never returns phone numbers, openids or device fingerprints."`
}

// FirstActivatorsRes is the response for the public first-activator wall.
type FirstActivatorsRes struct {
	List []*FirstActivatorItem `json:"list" dc:"First activators ordered by activation time ascending, capped at 120" eg:"[]"`
}

// FirstActivatorItem describes one first activator on the public wall. It carries
// only the player's nickname and identity label plus the activated cattle and the
// activation time; no privacy fields are exposed.
type FirstActivatorItem struct {
	Seq          int    `json:"seq" dc:"1-based global position on the first-activator wall" eg:"1"`
	UserId       int64  `json:"userId" dc:"Player ID" eg:"1"`
	Nickname     string `json:"nickname" dc:"Player nickname; empty when unset" eg:"川农牛仔"`
	IdentityType string `json:"identityType" dc:"Player identity label (student/alumni/teacher/friend)" eg:"student"`
	NiuId        int64  `json:"niuId" dc:"Activated cattle ID" eg:"7"`
	NiuName      string `json:"niuName" dc:"Activated cattle name; empty when missing" eg:"信工铁牛"`
	NiuCode      string `json:"niuCode" dc:"Activated cattle code; empty when missing" eg:"NIU-007"`
	ActivatedAt  *int64 `json:"activatedAt" dc:"Activation time as Unix milliseconds; null when unset" eg:"1717488000000"`
}

// HighlightsReq is the request for the public campus-history highlights.
type HighlightsReq struct {
	g.Meta `path:"/plugins/sicau-niu/wall/highlights" method:"get" tags:"Sicau Niu Wall" summary:"公共校史亮点" dc:"Return a bounded sample of campus-history highlights: cattle history cards and enabled campus-history quotes for the H5 memorial wall to render. This endpoint is public and requires no authentication."`
}

// HighlightsRes is the response for the public campus-history highlights.
type HighlightsRes struct {
	Cards  []*HighlightCard  `json:"cards" dc:"Bounded sample of campus-history cards" eg:"[]"`
	Quotes []*HighlightQuote `json:"quotes" dc:"Bounded sample of enabled campus-history quotes" eg:"[]"`
}

// HighlightCard describes one campus-history card on the public wall.
type HighlightCard struct {
	Id        int64  `json:"id" dc:"Card ID" eg:"1"`
	NiuId     int64  `json:"niuId" dc:"Cattle ID the card belongs to" eg:"7"`
	NiuName   string `json:"niuName" dc:"Cattle name; empty when missing" eg:"信工铁牛"`
	Category  string `json:"category" dc:"Card category" eg:"history"`
	Title     string `json:"title" dc:"Card title" eg:"建校120周年"`
	Content   string `json:"content" dc:"Card content; may be empty" eg:"……"`
	ImagePath string `json:"imagePath" dc:"Card image path; empty when none" eg:"/uploads/sicau-niu/card-1.png"`
}

// HighlightQuote describes one enabled campus-history quote on the public wall.
type HighlightQuote struct {
	Id      int64  `json:"id" dc:"Quote ID" eg:"1"`
	Content string `json:"content" dc:"Quote content" eg:"川农大精神:爱国敬业、艰苦奋斗、团结协作、求实创新"`
}

// ConfigReq is the request for the public memorial-wall configuration.
type ConfigReq struct {
	g.Meta `path:"/plugins/sicau-niu/wall/config" method:"get" tags:"Sicau Niu Wall" summary:"公共纪念墙配置" dc:"Return the public memorial-wall configuration: the return-to-mini-program URL the H5 wall links to. This endpoint is public and requires no authentication and exposes no privacy detail. The URL is empty when unconfigured, in which case the H5 hides the back-to-mini-program entry."`
}

// ConfigRes is the response for the public memorial-wall configuration.
type ConfigRes struct {
	MiniappURL string `json:"miniappUrl" dc:"Return-to-mini-program URL/scheme for the H5 wall; empty when unconfigured" eg:"weixin://dl/business/?t=XXXX"`
}

// StatsReq is the request for the public activity stats.
type StatsReq struct {
	g.Meta `path:"/plugins/sicau-niu/wall/stats" method:"get" tags:"Sicau Niu Wall" summary:"公共活动统计" dc:"Return public activity statistics aggregated on the database side: the number of activated cattle, the total cattle count, the first-activator count and the participating-player count. This endpoint is public and requires no authentication and exposes no privacy detail."`
}

// StatsRes is the response for the public activity stats.
type StatsRes struct {
	ActivatedNiuCount   int64 `json:"activatedNiuCount" dc:"Number of cattle whose status is active" eg:"42"`
	TotalNiuCount       int64 `json:"totalNiuCount" dc:"Total cattle count" eg:"120"`
	FirstActivatorCount int64 `json:"firstActivatorCount" dc:"Number of first activators (is_first activations)" eg:"42"`
	PlayerCount         int64 `json:"playerCount" dc:"Number of participating players" eg:"3500"`
}
