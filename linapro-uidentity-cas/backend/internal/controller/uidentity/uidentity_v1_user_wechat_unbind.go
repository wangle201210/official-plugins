package uidentity

import (
	"context"

	"lina-plugin-linapro-uidentity-cas/backend/api/uidentity/v1"
)

// UserWechatUnbind unbinds one runtime account Wechat union ID.
func (c *ControllerV1) UserWechatUnbind(ctx context.Context, req *v1.UserWechatUnbindReq) (res *v1.UserWechatUnbindRes, err error) {
	number, err := c.runtimeNumber(ctx)
	if err != nil {
		return nil, err
	}
	if err := c.uidentitySvc.UnbindRuntimeWechat(ctx, number); err != nil {
		return nil, err
	}
	return &v1.UserMutationRes{}, nil
}
