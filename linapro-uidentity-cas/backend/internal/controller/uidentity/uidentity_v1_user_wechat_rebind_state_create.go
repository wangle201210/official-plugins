// This file handles logged-in Wechat rebind state creation requests.

package uidentity

import (
	"context"

	v1 "lina-plugin-linapro-uidentity-cas/backend/api/uidentity/v1"
	uidentitysvc "lina-plugin-linapro-uidentity-cas/backend/internal/service/uidentity"
)

// UserWechatRebindStateCreate creates a short-lived rebind state.
func (c *ControllerV1) UserWechatRebindStateCreate(ctx context.Context, req *v1.UserWechatRebindStateCreateReq) (res *v1.UserWechatRebindStateCreateRes, err error) {
	out, err := c.uidentitySvc.CreateRuntimeWechatRebindState(ctx, uidentitysvc.WechatRebindStateInput{
		Number:   req.Number,
		Callback: req.Callback,
	})
	if err != nil {
		return nil, err
	}
	return &v1.UserWechatRebindStateCreateRes{
		State:     out.State,
		Status:    out.Status,
		Url:       out.URL,
		ExpiredAt: out.ExpiredAt,
	}, nil
}
