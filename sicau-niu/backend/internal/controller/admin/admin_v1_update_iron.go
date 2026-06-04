// admin_v1_update_iron.go implements the operator update-iron-cow handler.

package admin

import (
	"context"

	"lina-plugin-sicau-niu/backend/api/admin/v1"
	cattlesvc "lina-plugin-sicau-niu/backend/internal/service/cattle"
)

// UpdateIron updates one iron-cow registration.
func (c *ControllerV1) UpdateIron(ctx context.Context, req *v1.UpdateIronReq) (res *v1.UpdateIronRes, err error) {
	err = c.cattleSvc.UpdateIron(ctx, req.Id, &cattlesvc.IronMutateInput{
		Code:   req.Code,
		Name:   req.Name,
		Remark: req.Remark,
	})
	if err != nil {
		return nil, err
	}
	return &v1.UpdateIronRes{}, nil
}
