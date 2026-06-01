package uidentity

import (
	"context"

	"lina-plugin-linapro-uidentity-cas/backend/api/uidentity/v1"
	uidentitysvc "lina-plugin-linapro-uidentity-cas/backend/internal/service/uidentity"
)

// ActivationPassword sets activation password.
func (c *ControllerV1) ActivationPassword(ctx context.Context, req *v1.ActivationPasswordReq) (res *v1.ActivationPasswordRes, err error) {
	out, err := c.uidentitySvc.SetActivationPassword(ctx, uidentitysvc.ActivationPasswordInput{
		ChallengeID: req.ChallengeId,
		Password:    req.Password,
	})
	if err != nil {
		return nil, err
	}
	return &v1.ActivationStepRes{ChallengeId: out.ChallengeID, Success: out.Success}, nil
}
