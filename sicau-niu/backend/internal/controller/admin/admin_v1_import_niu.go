// admin_v1_import_niu.go implements bounded cattle batch import.
package admin

import (
	"context"

	"lina-plugin-sicau-niu/backend/api/admin/v1"
	cattlesvc "lina-plugin-sicau-niu/backend/internal/service/cattle"
)

func (c *ControllerV1) ImportNiu(ctx context.Context, req *v1.ImportNiuReq) (*v1.ImportNiuRes, error) {
	items := make([]*cattlesvc.NiuMutateInput, 0, len(req.Items))
	for _, item := range req.Items {
		if item == nil {
			items = append(items, nil)
			continue
		}
		items = append(items, &cattlesvc.NiuMutateInput{
			Code: item.Code, NiuType: item.NiuType, SpecialSubtype: item.SpecialSubtype, Name: item.Name,
			CollegeId: item.CollegeId, Lat: item.Lat, Lng: item.Lng, OnlineAt: item.OnlineAt,
			VisibleWeekdays: item.VisibleWeekdays, VisibleStart: item.VisibleStart, VisibleEnd: item.VisibleEnd,
		})
	}
	out, err := c.cattleSvc.ImportNiu(ctx, &cattlesvc.ImportNiuInput{Items: items, Overwrite: req.Overwrite})
	if err != nil {
		return nil, err
	}
	return &v1.ImportNiuRes{Created: out.Created, Updated: out.Updated, Skipped: out.Skipped}, nil
}
