package uidentity

import (
	"context"

	"lina-plugin-linapro-uidentity-cas/backend/api/uidentity/v1"
)

// AccountPasswordSelfReset resets password by a verified challenge.
func (c *ControllerV1) AccountPasswordSelfReset(ctx context.Context, req *v1.AccountPasswordSelfResetReq) (res *v1.AccountPasswordSelfResetRes, err error) {
	if err := c.uidentitySvc.ResetPasswordByChallenge(ctx, req.ChallengeId, req.NewPassword); err != nil {
		return nil, err
	}
	return &v1.AccountPasswordSelfResetRes{}, nil
}
