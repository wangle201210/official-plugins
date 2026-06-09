// admin_v1_update_niu.go implements the operator update-cattle handler.

package admin

import (
	"context"

	"lina-plugin-sicau-niu/backend/api/admin/v1"
	cattlesvc "lina-plugin-sicau-niu/backend/internal/service/cattle"
)

// UpdateNiu updates one cattle.
func (c *ControllerV1) UpdateNiu(ctx context.Context, req *v1.UpdateNiuReq) (res *v1.UpdateNiuRes, err error) {
	err = c.cattleSvc.UpdateNiu(ctx, req.Id, &cattlesvc.NiuMutateInput{
		Code:            req.Code,
		NiuType:         req.NiuType,
		SpecialSubtype:  req.SpecialSubtype,
		Name:            req.Name,
		CollegeId:       req.CollegeId,
		Lat:             req.Lat,
		Lng:             req.Lng,
		OnlineAt:        req.OnlineAt,
		VisibleWeekdays: req.VisibleWeekdays,
		VisibleStart:    req.VisibleStart,
		VisibleEnd:      req.VisibleEnd,
	})
	if err != nil {
		return nil, err
	}
	return &v1.UpdateNiuRes{}, nil
}
