// admin_v1_create_niu.go implements the operator create-cattle handler.

package admin

import (
	"context"

	"lina-plugin-sicau-niu/backend/api/admin/v1"
	cattlesvc "lina-plugin-sicau-niu/backend/internal/service/cattle"
)

// CreateNiu creates one cattle.
func (c *ControllerV1) CreateNiu(ctx context.Context, req *v1.CreateNiuReq) (res *v1.CreateNiuRes, err error) {
	id, err := c.cattleSvc.CreateNiu(ctx, &cattlesvc.NiuMutateInput{
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
	return &v1.CreateNiuRes{Id: id}, nil
}
