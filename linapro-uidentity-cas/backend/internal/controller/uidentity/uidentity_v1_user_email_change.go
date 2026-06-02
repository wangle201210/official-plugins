package uidentity

import (
	"context"

	"lina-plugin-linapro-uidentity-cas/backend/api/uidentity/v1"
)

// UserEmailChange changes one runtime account email.
func (c *ControllerV1) UserEmailChange(ctx context.Context, req *v1.UserEmailChangeReq) (res *v1.UserEmailChangeRes, err error) {
	number, err := c.runtimeNumber(ctx)
	if err != nil {
		return nil, err
	}
	if err := c.uidentitySvc.ChangeRuntimeEmail(ctx, number, req.Email); err != nil {
		return nil, err
	}
	return &v1.UserMutationRes{}, nil
}
