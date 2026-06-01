package uidentity

import (
	"context"

	"lina-plugin-linapro-uidentity-cas/backend/api/uidentity/v1"
	uidentitysvc "lina-plugin-linapro-uidentity-cas/backend/internal/service/uidentity"
)

// UserAppRoleUpdate updates delegated role expiration.
func (c *ControllerV1) UserAppRoleUpdate(ctx context.Context, req *v1.UserAppRoleUpdateReq) (res *v1.UserAppRoleUpdateRes, err error) {
	if err := c.uidentitySvc.UpdateRuntimeAppRole(ctx, uidentitysvc.UserAppRoleUpdateInput{
		Number:   req.Number,
		ID:       req.Id,
		ExpireAt: req.ExpireAt,
	}); err != nil {
		return nil, err
	}
	return &v1.UserMutationRes{}, nil
}
