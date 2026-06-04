// admin_v1_list_iron.go implements the operator iron-cow list handler.

package admin

import (
	"context"

	"lina-plugin-sicau-niu/backend/api/admin/v1"
	cattlesvc "lina-plugin-sicau-niu/backend/internal/service/cattle"
)

// ListIron returns one DB-side paged iron-cow list page.
func (c *ControllerV1) ListIron(ctx context.Context, req *v1.ListIronReq) (res *v1.ListIronRes, err error) {
	out, err := c.cattleSvc.ListIron(ctx, &cattlesvc.ListIronInput{
		Keyword:  req.Keyword,
		PageNum:  req.PageNum,
		PageSize: req.PageSize,
	})
	if err != nil {
		return nil, err
	}
	items := make([]*v1.IronItem, 0, len(out.List))
	for _, item := range out.List {
		items = append(items, &v1.IronItem{
			Id:        item.Id,
			Code:      item.Code,
			Name:      item.Name,
			LastLat:   item.LastLat,
			LastLng:   item.LastLng,
			LocatedAt: item.LocatedAt,
			Remark:    item.Remark,
			CreatedAt: item.CreatedAt,
			UpdatedAt: item.UpdatedAt,
		})
	}
	return &v1.ListIronRes{List: items, Total: out.Total}, nil
}
