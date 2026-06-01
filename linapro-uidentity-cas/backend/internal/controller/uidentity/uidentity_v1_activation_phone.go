package uidentity

import (
	"context"

	"lina-plugin-linapro-uidentity-cas/backend/api/uidentity/v1"
	uidentitysvc "lina-plugin-linapro-uidentity-cas/backend/internal/service/uidentity"
)

// ActivationPhone binds activation phone and marks the account active.
func (c *ControllerV1) ActivationPhone(ctx context.Context, req *v1.ActivationPhoneReq) (res *v1.ActivationPhoneRes, err error) {
	out, err := c.uidentitySvc.SetActivationPhone(ctx, uidentitysvc.ActivationPhoneInput{
		ChallengeID: req.ChallengeId,
		Phone:       req.Phone,
		Code:        req.Code,
	})
	if err != nil {
		return nil, err
	}
	return &v1.ActivationStepRes{ChallengeId: out.ChallengeID, Success: out.Success}, nil
}
