// admin_v1_create_honor.go implements the operator create-honor handler.

package admin

import (
	"context"

	"lina-plugin-sicau-niu/backend/api/admin/v1"
	honorsvc "lina-plugin-sicau-niu/backend/internal/service/honor"
)

// CreateHonor creates one honor definition.
func (c *ControllerV1) CreateHonor(ctx context.Context, req *v1.CreateHonorReq) (res *v1.CreateHonorRes, err error) {
	id, err := c.honorSvc.Create(ctx, &honorsvc.MutateInput{
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
	return &v1.CreateHonorRes{Id: id}, nil
}
