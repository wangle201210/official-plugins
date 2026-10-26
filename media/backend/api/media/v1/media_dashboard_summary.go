// This file defines the platform-wide active stream and session count endpoint.

package v1

import "github.com/gogf/gf/v2/frame/g"

// GetDashboardSummaryReq requests platform monitoring totals without list filters.
type GetDashboardSummaryReq struct {
	g.Meta `path:"/media/dashboard/summary" method:"get" tags:"媒体数据看板" summary:"查询活跃流和会话总数" dc:"查询全平台未关闭的流和会话总量，无需参数，不按租户筛选；以close_time为空判定活跃，不受列表10000条上限影响，不按心跳超时剔除。沿用媒体管理鉴权。" permission:"media:management:query"`
}

// GetDashboardSummaryRes contains exact counts from the latest lifecycle projections.
type GetDashboardSummaryRes struct {
	ActiveStreamCount  int64 `json:"active_stream_count" dc:"全平台活跃流总数，统计close_time为空的流记录；无记录时为0" eg:"128"`
	ActiveSessionCount int64 `json:"active_session_count" dc:"全平台活跃会话总数，统计close_time为空的会话记录；无记录时为0" eg:"1024"`
}
