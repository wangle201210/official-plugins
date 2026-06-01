// This file handles activation Wechat state creation requests.

package uidentity

import (
	"context"

	v1 "lina-plugin-linapro-uidentity-cas/backend/api/uidentity/v1"
	uidentitysvc "lina-plugin-linapro-uidentity-cas/backend/internal/service/uidentity"
)

// ActivationWechatStateCreate creates a short-lived activation Wechat state.
func (c *ControllerV1) ActivationWechatStateCreate(ctx context.Context, req *v1.ActivationWechatStateCreateReq) (res *v1.ActivationWechatStateCreateRes, err error) {
	out, err := c.uidentitySvc.CreateActivationWechatState(ctx, uidentitysvc.ActivationWechatStateInput{
		ChallengeID: req.ChallengeId,
		Callback:    req.Callback,
	})
	if err != nil {
		return nil, err
	}
	return &v1.ActivationWechatStateCreateRes{
		State:  out.State,
		Status: out.Status,
		Url:    out.URL,
	}, nil
}
