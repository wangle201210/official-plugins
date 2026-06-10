// This file declares media strategy device binding lookup DTOs.

package v1

import "github.com/gogf/gf/v2/frame/g"

// ListStrategyDeviceBindingsReq defines the request for querying device bindings by strategy ID.
type ListStrategyDeviceBindingsReq struct {
	g.Meta     `path:"/media/strategies/{strategyId}/device-bindings" method:"get" tags:"策略绑定" summary:"按策略查询设备绑定列表" dc:"分页查询绑定到指定媒体策略的设备策略绑定关系，只读取设备策略绑定表。" permission:"media:management:query"`
	StrategyId int64 `json:"strategyId" v:"required|min:1#策略ID不能为空|策略ID必须大于0" dc:"策略ID" eg:"1"`
	PageNum    int   `json:"pageNum" d:"1" v:"min:1" dc:"页码" eg:"1"`
	PageSize   int   `json:"pageSize" d:"10" v:"min:1|max:10000" dc:"每页条数，最大10000" eg:"10"`
}

// ListStrategyDeviceBindingsRes defines the response for querying device bindings by strategy ID.
type ListStrategyDeviceBindingsRes struct {
	List  []*StrategyDeviceBindingItem `json:"list" dc:"设备策略绑定列表" eg:"[]"`
	Total int                          `json:"total" dc:"匹配总数" eg:"1"`
}

// StrategyDeviceBindingItem defines one strategy-scoped device binding row.
type StrategyDeviceBindingItem struct {
	RowKey     string `json:"rowKey" dc:"前端表格行唯一键" eg:"device:34020000001320000001"`
	DeviceId   string `json:"deviceId" dc:"设备国标ID" eg:"34020000001320000001"`
	StrategyId int64  `json:"strategyId" dc:"策略ID" eg:"1"`
}
