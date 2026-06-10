// This file implements the strategy-scoped device binding list controller endpoint.

package media

import (
	"context"

	"lina-plugin-media/backend/api/media/v1"
	mediasvc "lina-plugin-media/backend/internal/service/media"
)

// ListStrategyDeviceBindings returns paged device bindings for one media strategy.
func (c *ControllerV1) ListStrategyDeviceBindings(ctx context.Context, req *v1.ListStrategyDeviceBindingsReq) (res *v1.ListStrategyDeviceBindingsRes, err error) {
	out, err := c.mediaSvc.ListStrategyDeviceBindings(ctx, mediasvc.ListStrategyDeviceBindingsInput{
		StrategyId: req.StrategyId,
		PageNum:    req.PageNum,
		PageSize:   req.PageSize,
	})
	if err != nil {
		return nil, err
	}
	items := make([]*v1.StrategyDeviceBindingItem, 0, len(out.List))
	for _, item := range out.List {
		items = append(items, &v1.StrategyDeviceBindingItem{
			RowKey:     item.RowKey,
			DeviceId:   item.DeviceId,
			StrategyId: item.StrategyId,
		})
	}
	return &v1.ListStrategyDeviceBindingsRes{List: items, Total: out.Total}, nil
}
