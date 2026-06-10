package uidentity

import (
	"context"

	"lina-plugin-linapro-uidentity-cas/backend/api/uidentity/v1"
	uidentitysvc "lina-plugin-linapro-uidentity-cas/backend/internal/service/uidentity"
)

// UserAppRoles returns delegated app roles for one runtime account.
func (c *ControllerV1) UserAppRoles(ctx context.Context, req *v1.UserAppRolesReq) (res *v1.UserAppRolesRes, err error) {
	number, err := c.runtimeNumber(ctx)
	if err != nil {
		return nil, err
	}
	out, err := c.uidentitySvc.ListRuntimeAppRoles(ctx, uidentitysvc.UserAppRoleListInput{
		Number:             number,
		EmpoweredAccountID: req.EmpoweredAccountId,
		PageNum:            req.PageNum,
		PageSize:           req.PageSize,
	})
	if err != nil {
		return nil, err
	}
	return &v1.UserAppRolesRes{List: toAPIRecords(out.List), Total: out.Total}, nil
}
