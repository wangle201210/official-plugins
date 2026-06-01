package uidentity

import (
	"context"

	"lina-plugin-linapro-uidentity-cas/backend/api/uidentity/v1"
)

// WechatLoginQRResult reads and consumes one terminal QR login result.
func (c *ControllerV1) WechatLoginQRResult(ctx context.Context, req *v1.WechatLoginQRResultReq) (res *v1.WechatLoginQRResultRes, err error) {
	out, err := c.uidentitySvc.GetWechatLoginQRResult(ctx, req.State)
	if err != nil {
		return nil, err
	}
	return toAPIWechatLoginResult(out), nil
}
