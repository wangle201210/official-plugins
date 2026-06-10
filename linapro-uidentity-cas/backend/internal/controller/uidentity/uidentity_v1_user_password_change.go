package uidentity

import (
	"context"

	"lina-plugin-linapro-uidentity-cas/backend/api/uidentity/v1"
)

// UserPasswordChange changes one runtime account password.
func (c *ControllerV1) UserPasswordChange(ctx context.Context, req *v1.UserPasswordChangeReq) (res *v1.UserPasswordChangeRes, err error) {
	number, err := c.runtimeNumber(ctx)
	if err != nil {
		return nil, err
	}
	// The old third-party password change verified an image captcha besides
	// the API signature, accepting the rotating common pass as the code.
	if err := c.uidentitySvc.VerifyLoginCaptcha(ctx, req.Code, req.UUID); err != nil {
		return nil, err
	}
	if err := c.uidentitySvc.ChangeRuntimePassword(ctx, number, req.NewPassword); err != nil {
		return nil, err
	}
	return &v1.UserMutationRes{}, nil
}
