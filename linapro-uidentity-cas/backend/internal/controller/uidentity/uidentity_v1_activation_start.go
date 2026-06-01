package uidentity

import (
	"context"

	"lina-plugin-linapro-uidentity-cas/backend/api/uidentity/v1"
	uidentitysvc "lina-plugin-linapro-uidentity-cas/backend/internal/service/uidentity"
)

// ActivationStart starts an account activation challenge.
func (c *ControllerV1) ActivationStart(ctx context.Context, req *v1.ActivationStartReq) (res *v1.ActivationStartRes, err error) {
	out, err := c.uidentitySvc.StartActivation(ctx, uidentitysvc.ActivationStartInput{
		Number: req.Number,
		Name:   req.Name,
		Idcard: req.Idcard,
	})
	if err != nil {
		return nil, err
	}
	return &v1.ActivationStartRes{ChallengeId: out.ChallengeID, NeedFace: out.NeedFace, Status: out.Status}, nil
}
