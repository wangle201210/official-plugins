// admin_v1_delete_iron.go implements the operator delete-iron-cow handler.

package admin

import (
	"context"

	"lina-plugin-sicau-niu/backend/api/admin/v1"
)

// DeleteIron soft-deletes one iron-cow registration.
func (c *ControllerV1) DeleteIron(ctx context.Context, req *v1.DeleteIronReq) (res *v1.DeleteIronRes, err error) {
	err = c.cattleSvc.DeleteIron(ctx, req.Id)
	if err != nil {
		return nil, err
	}
	return &v1.DeleteIronRes{}, nil
}
