package uidentity

import (
	"context"

	"lina-plugin-linapro-uidentity-cas/backend/api/uidentity/v1"
	uidentitysvc "lina-plugin-linapro-uidentity-cas/backend/internal/service/uidentity"
)

// WechatLoginQR creates a short-lived Wechat QR login state.
func (c *ControllerV1) WechatLoginQR(ctx context.Context, req *v1.WechatLoginQRReq) (res *v1.WechatLoginQRRes, err error) {
	out, err := c.uidentitySvc.CreateWechatLoginQR(ctx, uidentitysvc.WechatLoginQRInput{
		ClientID: req.ClientId,
		Callback: req.Callback,
	})
	if err != nil {
		return nil, err
	}
	return &v1.WechatLoginQRRes{
		State:     out.State,
		Url:       out.URL,
		ExpiredAt: out.ExpiredAt,
	}, nil
}
