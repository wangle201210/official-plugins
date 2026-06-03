// admin_v1_get_honor.go implements the operator get-honor-detail handler.

package admin

import (
	"context"

	"lina-plugin-sicau-niu/backend/api/admin/v1"
)

// GetHonor returns one honor-definition detail by ID.
func (c *ControllerV1) GetHonor(ctx context.Context, req *v1.GetHonorReq) (res *v1.GetHonorRes, err error) {
	item, err := c.honorSvc.Get(ctx, req.Id)
	if err != nil {
		return nil, err
	}
	return &v1.GetHonorRes{HonorItem: toHonorItem(item)}, nil
}
