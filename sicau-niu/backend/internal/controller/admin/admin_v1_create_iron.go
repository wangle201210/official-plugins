// admin_v1_create_iron.go implements the operator create-iron-cow handler.

package admin

import (
	"context"

	"lina-plugin-sicau-niu/backend/api/admin/v1"
	cattlesvc "lina-plugin-sicau-niu/backend/internal/service/cattle"
)

// CreateIron registers one iron-cow identifier.
func (c *ControllerV1) CreateIron(ctx context.Context, req *v1.CreateIronReq) (res *v1.CreateIronRes, err error) {
	id, err := c.cattleSvc.CreateIron(ctx, &cattlesvc.IronMutateInput{
		Code:   req.Code,
		Name:   req.Name,
		Remark: req.Remark,
	})
	if err != nil {
		return nil, err
	}
	return &v1.CreateIronRes{Id: id}, nil
}
