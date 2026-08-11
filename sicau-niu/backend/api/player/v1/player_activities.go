package v1

import "github.com/gogf/gf/v2/frame/g"

type ActivitiesReq struct {
	g.Meta `path:"/plugins/sicau-niu/player/activities" method:"get" tags:"寻牛小程序" summary:"查询本人好友动态" dc:"Return recent steal and gift events affecting the authenticated player. Events are bounded, newest first and batch-assembled without N+1 queries."`
}

type ActivitiesRes struct {
	List []*ActivityItem `json:"list" dc:"Recent activity rows ordered newest first and bounded by the server" eg:"[]"`
}

type ActivityItem struct {
	Id         string `json:"id" dc:"Stable event identifier" eg:"steal:1"`
	Name       string `json:"name" dc:"Other player's nickname" eg:"川农同学"`
	Action     string `json:"action" dc:"Display action" eg:"领走了你的"`
	Amount     string `json:"amount" dc:"Display amount" eg:"2捆"`
	Target     string `json:"target" dc:"Display target" eg:"苜蓿草"`
	OccurredAt *int64 `json:"occurredAt" dc:"Event time as Unix timestamp in milliseconds" eg:"1776333600000"`
	Type       string `json:"type" dc:"Mini-program event type: steal or help" eg:"steal"`
}
