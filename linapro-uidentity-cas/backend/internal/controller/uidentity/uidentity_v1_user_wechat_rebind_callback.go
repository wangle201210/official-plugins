// This file handles external Wechat rebind callback completion requests.

package uidentity

import (
	"context"

	v1 "lina-plugin-linapro-uidentity-cas/backend/api/uidentity/v1"
	uidentitysvc "lina-plugin-linapro-uidentity-cas/backend/internal/service/uidentity"
)

// UserWechatRebindCallback records a Wechat rebind callback result.
func (c *ControllerV1) UserWechatRebindCallback(ctx context.Context, req *v1.UserWechatRebindCallbackReq) (res *v1.UserWechatRebindCallbackRes, err error) {
	out, err := c.uidentitySvc.CompleteRuntimeWechatRebind(ctx, uidentitysvc.WechatRebindCallbackInput{
		State:    req.State,
		UnionID:  req.UnionId,
		Code:     req.Code,
		Callback: req.Callback,
	})
	if err != nil {
		return nil, err
	}
	return toAPIWechatRebindState(out), nil
}
