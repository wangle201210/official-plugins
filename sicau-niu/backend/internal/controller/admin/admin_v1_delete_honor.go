// admin_v1_delete_honor.go implements the operator delete-honor handler.

package admin

import (
	"context"

	"lina-plugin-sicau-niu/backend/api/admin/v1"
)

// DeleteHonor soft-deletes one honor definition.
func (c *ControllerV1) DeleteHonor(ctx context.Context, req *v1.DeleteHonorReq) (res *v1.DeleteHonorRes, err error) {
	if err = c.honorSvc.Delete(ctx, req.Id); err != nil {
		return nil, err
	}
	return &v1.DeleteHonorRes{}, nil
}
