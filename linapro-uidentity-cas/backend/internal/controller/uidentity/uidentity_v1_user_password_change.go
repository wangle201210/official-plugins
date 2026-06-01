package uidentity

import (
	"context"

	"lina-plugin-linapro-uidentity-cas/backend/api/uidentity/v1"
)

// UserPasswordChange changes one runtime account password.
func (c *ControllerV1) UserPasswordChange(ctx context.Context, req *v1.UserPasswordChangeReq) (res *v1.UserPasswordChangeRes, err error) {
	if err := c.uidentitySvc.ChangeRuntimePassword(ctx, req.Number, req.NewPassword); err != nil {
		return nil, err
	}
	return &v1.UserMutationRes{}, nil
}
