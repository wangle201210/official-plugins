// This file defines operator-facing cloud-moving report audit APIs.

package v1

import "github.com/gogf/gf/v2/frame/g"

type ListTransportReportsReq struct {
	g.Meta       `path:"/plugins/sicau-niu/admin/transport-reports" method:"get" tags:"寻牛运营" summary:"分页查询云搬牛上报事实" dc:"Audit successful position report facts with bounded pagination. Raw coordinates require the transport audit permission." permission:"sicau-niu:transport:audit"`
	PageNum      int    `json:"pageNum" dc:"Page number; defaults to 1" eg:"1"`
	PageSize     int    `json:"pageSize" dc:"Page size; defaults to 20 and is capped at 100" eg:"20"`
	TeamId       int64  `json:"teamId" dc:"Optional team ID filter" eg:"12"`
	UserId       int64  `json:"userId" dc:"Optional player ID filter" eg:"8"`
	ActivityDate string `json:"activityDate" dc:"Optional Beijing natural-day key" eg:"2026-08-12"`
}

type ListTransportReportsRes struct {
	List  []*AdminTransportReport `json:"list" dc:"Page of successful report facts" eg:"[]"`
	Total int                     `json:"total" dc:"Matched report count" eg:"1"`
}
