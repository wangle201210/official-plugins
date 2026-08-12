// This file defines operator-facing cloud-moving aggregate APIs.

package v1

import "github.com/gogf/gf/v2/frame/g"

type GetTransportStatsReq struct {
	g.Meta       `path:"/plugins/sicau-niu/admin/transport-stats" method:"get" tags:"寻牛运营" summary:"查询云搬牛统计" dc:"Aggregate cloud-moving team, report and contribution facts in the database. An optional Beijing day key narrows report metrics." permission:"sicau-niu:transport:list"`
	ActivityDate string `json:"activityDate" dc:"Optional Beijing natural-day key in YYYY-MM-DD" eg:"2026-08-12"`
}

type GetTransportStatsRes struct {
	EffectiveTeamCount int   `json:"effectiveTeamCount" dc:"Current effective team count" eg:"12"`
	InvalidTeamCount   int   `json:"invalidTeamCount" dc:"Historical invalid team count" eg:"3"`
	ActiveMemberCount  int   `json:"activeMemberCount" dc:"Current active membership count" eg:"30"`
	ReportCount        int   `json:"reportCount" dc:"Matched successful report count" eg:"128"`
	ContributionMeters int64 `json:"contributionMeters" dc:"Matched contribution meters" eg:"53210"`
}
