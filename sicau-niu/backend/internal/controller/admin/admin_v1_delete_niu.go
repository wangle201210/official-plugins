// admin_v1_delete_niu.go implements the operator delete-cattle handler.

package admin

import (
	"context"

	"lina-plugin-sicau-niu/backend/api/admin/v1"
)

// DeleteNiu soft-deletes one cattle and cascade soft-deletes its main card.
func (c *ControllerV1) DeleteNiu(ctx context.Context, req *v1.DeleteNiuReq) (res *v1.DeleteNiuRes, err error) {
	err = c.cattleSvc.DeleteNiu(ctx, req.Id)
	if err != nil {
		return nil, err
	}
	return &v1.DeleteNiuRes{}, nil
}
