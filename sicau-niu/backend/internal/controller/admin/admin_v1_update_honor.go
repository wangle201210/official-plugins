// admin_v1_update_honor.go implements the operator update-honor handler.

package admin

import (
	"context"

	"lina-plugin-sicau-niu/backend/api/admin/v1"
	honorsvc "lina-plugin-sicau-niu/backend/internal/service/honor"
)

// UpdateHonor updates one honor definition.
func (c *ControllerV1) UpdateHonor(ctx context.Context, req *v1.UpdateHonorReq) (res *v1.UpdateHonorRes, err error) {
	err = c.honorSvc.Update(ctx, req.Id, &honorsvc.MutateInput{
		HonorType:  req.HonorType,
		Code:       req.Code,
		Name:       req.Name,
		UnlockType: req.UnlockType,
		Threshold:  req.Threshold,
		Category:   req.Category,
		ImagePath:  req.ImagePath,
		Sort:       req.Sort,
	})
	if err != nil {
		return nil, err
	}
	return &v1.UpdateHonorRes{}, nil
}
