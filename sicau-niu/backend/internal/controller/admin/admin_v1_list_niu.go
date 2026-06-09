// admin_v1_list_niu.go implements the operator cattle list handler.

package admin

import (
	"context"

	"lina-plugin-sicau-niu/backend/api/admin/v1"
	cattlesvc "lina-plugin-sicau-niu/backend/internal/service/cattle"
)

// ListNiu returns one DB-side paged cattle list page with batch-assembled
// college names and card-binding flags.
func (c *ControllerV1) ListNiu(ctx context.Context, req *v1.ListNiuReq) (res *v1.ListNiuRes, err error) {
	out, err := c.cattleSvc.ListNiu(ctx, &cattlesvc.ListNiuInput{
		Keyword:  req.Keyword,
		NiuType:  req.NiuType,
		PageNum:  req.PageNum,
		PageSize: req.PageSize,
	})
	if err != nil {
		return nil, err
	}
	items := make([]*v1.NiuItem, 0, len(out.List))
	for _, item := range out.List {
		items = append(items, toNiuItem(item))
	}
	return &v1.ListNiuRes{List: items, Total: out.Total}, nil
}
