// This file defines player-facing cloud-moving contribution APIs.

package v1

import "github.com/gogf/gf/v2/frame/g"

type ReportTransportContributionReq struct {
	g.Meta    `path:"/plugins/sicau-niu/player/iron-transport/contributions" method:"post" tags:"寻牛小程序" summary:"上报云搬牛贡献" dc:"Persist one GCJ-02 position report fact for the current player's effective team. The first report contributes zero meters; later reports contribute the rounded Haversine distance from the previous successful point. Each player may report at most 12 times per Beijing natural day. Requires a valid player token."`
	TeamId    int64    `json:"teamId" v:"required|min:1" dc:"Current effective team ID" eg:"12"`
	Lat       *float64 `json:"lat" v:"required|between:-90,90" dc:"Current GCJ-02 latitude" eg:"30.7059"`
	Lng       *float64 `json:"lng" v:"required|between:-180,180" dc:"Current GCJ-02 longitude" eg:"103.8318"`
	SampledAt int64    `json:"sampledAt" v:"required|min:1" dc:"Client sample time as Unix milliseconds; audit-only and does not control daily quota" eg:"1776333600000"`
	RequestId string   `json:"requestId" v:"required|max-length:64" dc:"Player-scoped idempotency key" eg:"transport-report-1b2a3c4d"`
}

type ReportTransportContributionRes struct {
	Report *IronTransportReportResult `json:"report" dc:"Stable accepted report result" eg:"{}"`
}

type ListMyTransportReportsReq struct {
	g.Meta   `path:"/plugins/sicau-niu/player/iron-transport/contributions/mine" method:"get" tags:"寻牛小程序" summary:"分页查询本人云搬牛上报" dc:"Return the current player's own successful report facts, including raw coordinates, with bounded pagination. Requires a valid player token."`
	PageNum  int `json:"pageNum" dc:"Page number; defaults to 1" eg:"1"`
	PageSize int `json:"pageSize" dc:"Page size; defaults to 20 and is capped at 100" eg:"20"`
}

type ListMyTransportReportsRes struct {
	List  []*IronTransportReport `json:"list" dc:"Page of the player's report facts" eg:"[]"`
	Total int                    `json:"total" dc:"Matched report count" eg:"1"`
}
