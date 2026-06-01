package uidentity

import (
	"context"

	"lina-plugin-linapro-uidentity-cas/backend/api/uidentity/v1"
)

// AccountPasswordChallenge creates a self-service password reset challenge.
func (c *ControllerV1) AccountPasswordChallenge(ctx context.Context, req *v1.AccountPasswordChallengeReq) (res *v1.AccountPasswordChallengeRes, err error) {
	out, err := c.uidentitySvc.CreatePasswordChallenge(ctx, req.Number)
	if err != nil {
		return nil, err
	}
	return &v1.AccountPasswordChallengeRes{ChallengeId: out.ChallengeID, Status: out.Status}, nil
}
