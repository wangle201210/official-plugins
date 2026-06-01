package uidentity

import (
	"context"

	"lina-plugin-linapro-uidentity-cas/backend/api/uidentity/v1"
	uidentitysvc "lina-plugin-linapro-uidentity-cas/backend/internal/service/uidentity"
)

// WechatLoginCallback records one Wechat QR callback result.
func (c *ControllerV1) WechatLoginCallback(ctx context.Context, req *v1.WechatLoginCallbackReq) (res *v1.WechatLoginCallbackRes, err error) {
	out, err := c.uidentitySvc.CompleteWechatLoginQR(ctx, uidentitysvc.WechatLoginCallbackInput{
		State:    req.State,
		ClientID: req.ClientId,
		Code:     req.Code,
		UnionID:  req.UnionId,
		Callback: req.Callback,
	})
	if err != nil {
		return nil, err
	}
	return toAPIWechatLoginResult(out), nil
}
