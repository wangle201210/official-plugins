package uidentity

import (
	"context"

	"lina-plugin-linapro-uidentity-cas/backend/api/uidentity/v1"
)

// ActivationState returns the current activation state.
func (c *ControllerV1) ActivationState(ctx context.Context, req *v1.ActivationStateReq) (res *v1.ActivationStateRes, err error) {
	out, err := c.uidentitySvc.ActivationState(ctx, req.ChallengeId)
	if err != nil {
		return nil, err
	}
	return &v1.ActivationStateRes{
		ChallengeId: out.ChallengeID,
		Success:     out.Success,
		Status:      out.Status,
		Stage:       out.Stage,
	}, nil
}
