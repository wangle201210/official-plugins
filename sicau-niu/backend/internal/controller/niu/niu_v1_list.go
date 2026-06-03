// niu_v1_list.go projects the protected sample cattle list service output to the
// sicau-niu API response, converting service records into bounded response items.

package niu

import (
	"context"

	"lina-plugin-sicau-niu/backend/api/niu/v1"
	niusvc "lina-plugin-sicau-niu/backend/internal/service/niu"
)

// List returns the bounded sample cattle list for the plugin page.
func (c *ControllerV1) List(ctx context.Context, req *v1.ListReq) (res *v1.ListRes, err error) {
	out, err := c.niuSvc.List(ctx, &niusvc.ListInput{Keyword: req.Keyword})
	if err != nil {
		return nil, err
	}

	items := make([]*v1.CattleItem, 0, len(out.Items))
	for _, record := range out.Items {
		items = append(items, &v1.CattleItem{
			ID:        record.ID,
			Name:      record.Name,
			Breed:     record.Breed,
			WeightKg:  record.WeightKg,
			CreatedAt: record.CreatedAtMs,
		})
	}

	return &v1.ListRes{List: items, Total: out.Total}, nil
}
