// This file handles logged-in Wechat rebind state lookup requests.

package uidentity

import (
	"context"

	v1 "lina-plugin-linapro-uidentity-cas/backend/api/uidentity/v1"
	uidentitysvc "lina-plugin-linapro-uidentity-cas/backend/internal/service/uidentity"
)

// UserWechatRebindState returns the current rebind state.
func (c *ControllerV1) UserWechatRebindState(ctx context.Context, req *v1.UserWechatRebindStateReq) (res *v1.UserWechatRebindStateRes, err error) {
	out, err := c.uidentitySvc.GetRuntimeWechatRebindState(ctx, uidentitysvc.WechatRebindStateLookupInput{
		Number: req.Number,
		State:  req.State,
	})
	if err != nil {
		return nil, err
	}
	return toAPIWechatRebindState(out), nil
}
