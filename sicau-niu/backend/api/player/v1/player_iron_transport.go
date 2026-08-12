// This file defines the player-facing cloud-moving team and contribution API.

package v1

import "github.com/gogf/gf/v2/frame/g"

type IronTransportStateReq struct {
	g.Meta `path:"/plugins/sicau-niu/player/iron-transport/state" method:"get" tags:"寻牛小程序" summary:"查询云搬牛状态" dc:"Return at most 120 effective cloud-moving teams, the current player's team projection and today's report quota. Expired teams are settled before the response. Requires a valid player token."`
}

type IronTransportStateRes struct {
	*IronTransportState `json:",inline" dc:"Current cloud-moving state" eg:"{}"`
}

type IronTransportState struct {
	Explanation          string               `json:"explanation" dc:"Short cloud-moving rule explanation" eg:"加入一个团后，每天最多上报12次位置贡献。"`
	MaxEffectiveTeams    int                  `json:"maxEffectiveTeams" dc:"Maximum concurrent effective teams" eg:"120"`
	DailyReportLimit     int                  `json:"dailyReportLimit" dc:"Per-player Beijing-day report limit" eg:"12"`
	TodayReportCount     int                  `json:"todayReportCount" dc:"Current player's accepted report count today" eg:"2"`
	TodayReportRemaining int                  `json:"todayReportRemaining" dc:"Current player's remaining reports today" eg:"10"`
	Teams                []*IronTransportTeam `json:"teams" dc:"All effective teams, bounded to 120 and without nested member arrays" eg:"[]"`
	MyTeam               *IronTransportTeam   `json:"myTeam,omitempty" dc:"Current player's effective team; omitted when unbound" eg:"null"`
}

type IronTransportTeam struct {
	Id                      string `json:"id" dc:"Stable team ID" eg:"12"`
	Name                    string `json:"name" dc:"Team display name" eg:"川农云搬牛一团"`
	CreatorId               string `json:"creatorId" dc:"Creator player ID" eg:"8"`
	CreatorName             string `json:"creatorName" dc:"Creator nickname" eg:"川农校友"`
	MemberCount             int    `json:"memberCount" dc:"Current active member count" eg:"3"`
	TotalContributionMeters int64  `json:"totalContributionMeters" dc:"Team cumulative contribution meters" eg:"1250"`
	MyContributionMeters    int64  `json:"myContributionMeters" dc:"Requesting player's cumulative contribution in this team; zero when not a member" eg:"450"`
	HasReportBaseline       bool   `json:"hasReportBaseline" dc:"Whether the requesting player has a previous successful point in this team" eg:"true"`
	LastActiveAt            *int64 `json:"lastActiveAt" dc:"Latest successful create, join or report server commit time as Unix milliseconds" eg:"1776333600000"`
	CreatedAt               *int64 `json:"createdAt" dc:"Team creation time as Unix milliseconds" eg:"1776333600000"`
	Mine                    bool   `json:"mine" dc:"Whether the requesting player is currently bound to this team" eg:"true"`
}

type IronTransportMember struct {
	Id                 string `json:"id" dc:"Membership ID" eg:"21"`
	UserId             string `json:"userId" dc:"Player ID" eg:"8"`
	Name               string `json:"name" dc:"Public nickname" eg:"川农校友"`
	Avatar             string `json:"avatar" dc:"Public avatar URL or value" eg:"https://example.com/avatar.png"`
	Role               string `json:"role" dc:"Display role: creator or member" eg:"creator"`
	ContributionMeters int64  `json:"contributionMeters" dc:"Member cumulative contribution meters in this team" eg:"450"`
	JoinedAt           *int64 `json:"joinedAt" dc:"Membership join time as Unix milliseconds" eg:"1776333600000"`
}

type IronTransportReportResult struct {
	Id                   string   `json:"id" dc:"Report fact ID" eg:"31"`
	TeamId               string   `json:"teamId" dc:"Owning team ID" eg:"12"`
	StartLat             *float64 `json:"startLat,omitempty" dc:"Previous successful latitude; omitted on the first report" eg:"30.7058"`
	StartLng             *float64 `json:"startLng,omitempty" dc:"Previous successful longitude; omitted on the first report" eg:"103.8317"`
	EndLat               float64  `json:"endLat" dc:"Accepted current latitude" eg:"30.7059"`
	EndLng               float64  `json:"endLng" dc:"Accepted current longitude" eg:"103.8318"`
	ContributionMeters   int64    `json:"contributionMeters" dc:"Contribution from this report in rounded integer meters" eg:"15"`
	UserTotalMeters      int64    `json:"userTotalMeters" dc:"Member cumulative contribution after this report" eg:"450"`
	TeamTotalMeters      int64    `json:"teamTotalMeters" dc:"Team cumulative contribution after this report" eg:"1250"`
	TodayReportCount     int      `json:"todayReportCount" dc:"Accepted report count today" eg:"2"`
	TodayReportRemaining int      `json:"todayReportRemaining" dc:"Remaining report count today" eg:"10"`
	AcceptedAt           *int64   `json:"acceptedAt" dc:"Server acceptance time as Unix milliseconds" eg:"1776333600000"`
}

type IronTransportReport struct {
	*IronTransportReportResult `json:",inline"`
	ActivityDate               string `json:"activityDate" dc:"Beijing natural-day key" eg:"2026-08-12"`
	SampledAt                  *int64 `json:"sampledAt" dc:"Client sample time as Unix milliseconds" eg:"1776333600000"`
}
