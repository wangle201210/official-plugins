// This file declares media strategy tenant binding lookup DTOs.

package v1

import "github.com/gogf/gf/v2/frame/g"

// ListStrategyTenantBindingsReq defines the request for querying tenant bindings by strategy ID.
type ListStrategyTenantBindingsReq struct {
	g.Meta     `path:"/media/strategies/{strategyId}/tenant-bindings" method:"get" tags:"策略绑定" summary:"按策略查询租户绑定列表" dc:"分页查询绑定到指定媒体策略的租户策略绑定关系，只读取租户策略绑定表。" permission:"media:management:query"`
	StrategyId int64 `json:"strategyId" v:"required|min:1#策略ID不能为空|策略ID必须大于0" dc:"策略ID" eg:"1"`
	PageNum    int   `json:"pageNum" d:"1" v:"min:1" dc:"页码" eg:"1"`
	PageSize   int   `json:"pageSize" d:"10" v:"min:1|max:10000" dc:"每页条数，最大10000" eg:"10"`
}

// ListStrategyTenantBindingsRes defines the response for querying tenant bindings by strategy ID.
type ListStrategyTenantBindingsRes struct {
	List  []*StrategyTenantBindingItem `json:"list" dc:"租户策略绑定列表" eg:"[]"`
	Total int                          `json:"total" dc:"匹配总数" eg:"1"`
}

// StrategyTenantBindingItem defines one strategy-scoped tenant binding row.
type StrategyTenantBindingItem struct {
	RowKey     string `json:"rowKey" dc:"前端表格行唯一键" eg:"tenant:tenant-a"`
	TenantId   string `json:"tenantId" dc:"租户ID" eg:"tenant-a"`
	StrategyId int64  `json:"strategyId" dc:"策略ID" eg:"1"`
}
