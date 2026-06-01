// This file handles external activation Wechat callback completion requests.

package uidentity

import (
	"context"

	v1 "lina-plugin-linapro-uidentity-cas/backend/api/uidentity/v1"
	uidentitysvc "lina-plugin-linapro-uidentity-cas/backend/internal/service/uidentity"
)

// ActivationWechatCallback records an activation Wechat callback result.
func (c *ControllerV1) ActivationWechatCallback(ctx context.Context, req *v1.ActivationWechatCallbackReq) (res *v1.ActivationWechatCallbackRes, err error) {
	out, err := c.uidentitySvc.CompleteActivationWechat(ctx, uidentitysvc.ActivationWechatCallbackInput{
		State:    req.State,
		UnionID:  req.UnionId,
		Code:     req.Code,
		Callback: req.Callback,
	})
	if err != nil {
		return nil, err
	}
	return toAPIActivationWechatState(out), nil
}
