// This file implements the strategy-scoped tenant binding list controller endpoint.

package media

import (
	"context"

	"lina-plugin-media/backend/api/media/v1"
	mediasvc "lina-plugin-media/backend/internal/service/media"
)

// ListStrategyTenantBindings returns paged tenant bindings for one media strategy.
func (c *ControllerV1) ListStrategyTenantBindings(ctx context.Context, req *v1.ListStrategyTenantBindingsReq) (res *v1.ListStrategyTenantBindingsRes, err error) {
	out, err := c.mediaSvc.ListStrategyTenantBindings(ctx, mediasvc.ListStrategyTenantBindingsInput{
		StrategyId: req.StrategyId,
		PageNum:    req.PageNum,
		PageSize:   req.PageSize,
	})
	if err != nil {
		return nil, err
	}
	items := make([]*v1.StrategyTenantBindingItem, 0, len(out.List))
	for _, item := range out.List {
		items = append(items, &v1.StrategyTenantBindingItem{
			RowKey:     item.RowKey,
			TenantId:   item.TenantId,
			StrategyId: item.StrategyId,
		})
	}
	return &v1.ListStrategyTenantBindingsRes{List: items, Total: out.Total}, nil
}
