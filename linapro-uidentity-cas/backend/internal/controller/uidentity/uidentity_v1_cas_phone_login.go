package uidentity

import (
	"context"

	"lina-plugin-linapro-uidentity-cas/backend/api/uidentity/v1"
	uidentitysvc "lina-plugin-linapro-uidentity-cas/backend/internal/service/uidentity"
)

// CasPhoneLogin validates phone SMS login and returns CAS tickets.
func (c *ControllerV1) CasPhoneLogin(ctx context.Context, req *v1.CasPhoneLoginReq) (res *v1.CasPhoneLoginRes, err error) {
	out, err := c.uidentitySvc.LoginByPhone(ctx, uidentitysvc.PhoneLoginInput{
		ClientID: req.ClientId,
		Phone:    req.Phone,
		Code:     req.Code,
	})
	if err != nil {
		return nil, err
	}
	return toAPIRuntimeLogin(out), nil
}
