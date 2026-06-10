package uidentity

import (
	"context"

	"lina-plugin-linapro-uidentity-cas/backend/api/uidentity/v1"
	uidentitysvc "lina-plugin-linapro-uidentity-cas/backend/internal/service/uidentity"
)

// CasPasswordLogin validates account password login and returns CAS tickets.
func (c *ControllerV1) CasPasswordLogin(ctx context.Context, req *v1.CasPasswordLoginReq) (res *v1.CasPasswordLoginRes, err error) {
	// The old CAS password login verified a captcha first, accepting the
	// rotating common pass in place of the captcha code.
	if err := c.uidentitySvc.VerifyLoginCaptcha(ctx, req.Code, req.UUID); err != nil {
		return nil, err
	}
	out, err := c.uidentitySvc.LoginByPassword(ctx, uidentitysvc.PasswordLoginInput{
		ClientID: req.ClientId,
		Number:   req.Number,
		Password: req.Password,
	})
	if err != nil {
		return nil, err
	}
	return toAPIRuntimeLogin(out), nil
}
