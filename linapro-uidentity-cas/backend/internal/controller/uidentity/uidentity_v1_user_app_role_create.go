package uidentity

import (
	"context"

	"lina-plugin-linapro-uidentity-cas/backend/api/uidentity/v1"
	uidentitysvc "lina-plugin-linapro-uidentity-cas/backend/internal/service/uidentity"
)

// UserAppRoleCreate creates one delegated app role.
func (c *ControllerV1) UserAppRoleCreate(ctx context.Context, req *v1.UserAppRoleCreateReq) (res *v1.UserAppRoleCreateRes, err error) {
	id, err := c.uidentitySvc.CreateRuntimeAppRole(ctx, uidentitysvc.UserAppRoleCreateInput{
		Number:          req.Number,
		EmpoweredNumber: req.EmpoweredNumber,
		AppID:           req.AppId,
		ExpireAt:        req.ExpireAt,
	})
	if err != nil {
		return nil, err
	}
	return &v1.UserAppRoleCreateRes{Id: id}, nil
}
