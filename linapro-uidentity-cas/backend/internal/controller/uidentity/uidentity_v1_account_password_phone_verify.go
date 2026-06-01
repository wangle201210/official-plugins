package uidentity

import (
	"context"

	"lina-plugin-linapro-uidentity-cas/backend/api/uidentity/v1"
)

// AccountPasswordPhoneVerify verifies a password reset challenge by phone.
func (c *ControllerV1) AccountPasswordPhoneVerify(ctx context.Context, req *v1.AccountPasswordPhoneVerifyReq) (res *v1.AccountPasswordPhoneVerifyRes, err error) {
	challengeID, err := c.uidentitySvc.VerifyPasswordChallengePhone(ctx, req.ChallengeId, req.Phone, req.Code)
	if err != nil {
		return nil, err
	}
	return &v1.AccountPasswordPhoneVerifyRes{ChallengeId: challengeID}, nil
}
