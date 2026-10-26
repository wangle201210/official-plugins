// This file defines platform-wide alive node, active stream and session counts.

package v1

import "github.com/gogf/gf/v2/frame/g"

// GetDashboardSummaryReq requests platform monitoring totals without list filters.
type GetDashboardSummaryReq struct {
	g.Meta `path:"/media/dashboard/summary" method:"get" tags:"媒体数据看板" summary:"查询存活节点、活跃流和会话总数" dc:"查询全平台存活节点及running且未关闭的流、关联未关闭会话总量，无需参数，不按租户筛选。流须同时满足status=running和close_time IS NULL；会话没有status字段，按stream_id关联这些流且自身close_time IS NULL。非running流、已关闭流、无对应流的会话均不计入。存活节点独立统计最近3分钟内上报的节点，按服务端记录的updated_at判断，不看节点status，不要求承载运行中流；缺失、超时或未来上报时间不计入。不受10000条列表上限影响。沿用媒体管理鉴权，结果位于统一响应code、message、data的data内。" permission:"media:management:query"`
}

// GetDashboardSummaryRes contains exact counts from the latest lifecycle projections.
type GetDashboardSummaryRes struct {
	ActiveStreamCount  int64 `json:"active_stream_count" dc:"全平台running且close_time为空的流总数；无匹配记录时为0" eg:"128"`
	ActiveSessionCount int64 `json:"active_session_count" dc:"running且未关闭的流所关联的未关闭会话总数；通过stream_id匹配，孤立会话不计入，无匹配记录时为0" eg:"1024"`
	AliveNodeCount     int64 `json:"alive_node_count" dc:"全平台最近3分钟内上报过数据的节点数，按服务端节点记录updated_at落在当前时间前3分钟至当前时间的闭区间判断；不按status判断，不要求有运行中流。缺失、超时或未来更新时间不计入，空数据为0" eg:"14"`
}
