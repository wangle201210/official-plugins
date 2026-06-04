// admin_v1_list_honor.go implements the operator honor-definition list handler.

package admin

import (
	"context"

	"lina-plugin-sicau-niu/backend/api/admin/v1"
	honorsvc "lina-plugin-sicau-niu/backend/internal/service/honor"
)

// ListHonor returns one DB-side paged honor-definition list page.
func (c *ControllerV1) ListHonor(ctx context.Context, req *v1.ListHonorReq) (res *v1.ListHonorRes, err error) {
	out, err := c.honorSvc.List(ctx, &honorsvc.ListInput{
		Keyword:    req.Keyword,
		HonorType:  req.HonorType,
		UnlockType: req.UnlockType,
		PageNum:    req.PageNum,
		PageSize:   req.PageSize,
	})
	if err != nil {
		return nil, err
	}
	items := make([]*v1.HonorItem, 0, len(out.List))
	for _, item := range out.List {
		items = append(items, toHonorItem(item))
	}
	return &v1.ListHonorRes{List: items, Total: out.Total}, nil
}
