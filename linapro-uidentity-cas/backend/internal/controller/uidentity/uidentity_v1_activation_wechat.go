package uidentity

import (
	"context"

	"lina-plugin-linapro-uidentity-cas/backend/api/uidentity/v1"
	uidentitysvc "lina-plugin-linapro-uidentity-cas/backend/internal/service/uidentity"
)

// ActivationWechat binds Wechat union ID and marks the account active.
func (c *ControllerV1) ActivationWechat(ctx context.Context, req *v1.ActivationWechatReq) (res *v1.ActivationWechatRes, err error) {
	out, err := c.uidentitySvc.SetActivationWechat(ctx, uidentitysvc.ActivationWechatInput{
		ChallengeID: req.ChallengeId,
		UnionID:     req.UnionId,
	})
	if err != nil {
		return nil, err
	}
	return &v1.ActivationStepRes{ChallengeId: out.ChallengeID, Success: out.Success}, nil
}
