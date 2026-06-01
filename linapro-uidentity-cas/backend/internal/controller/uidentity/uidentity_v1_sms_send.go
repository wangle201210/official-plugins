package uidentity

import (
	"context"

	"lina-plugin-linapro-uidentity-cas/backend/api/uidentity/v1"
	uidentitysvc "lina-plugin-linapro-uidentity-cas/backend/internal/service/uidentity"
)

// SmsSend records one plugin-local SMS verification code.
func (c *ControllerV1) SmsSend(ctx context.Context, req *v1.SmsSendReq) (res *v1.SmsSendRes, err error) {
	out, err := c.uidentitySvc.SendSMSCode(ctx, uidentitysvc.SMSSendInput{
		Type:  req.Type,
		Phone: req.Phone,
	})
	if err != nil {
		return nil, err
	}
	return &v1.SmsSendRes{Id: out.ID}, nil
}
