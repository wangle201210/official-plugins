package uidentity

import (
	"context"

	"lina-plugin-linapro-uidentity-cas/backend/api/uidentity/v1"
	uidentitysvc "lina-plugin-linapro-uidentity-cas/backend/internal/service/uidentity"
)

// UserPhoneChange changes one runtime account phone.
func (c *ControllerV1) UserPhoneChange(ctx context.Context, req *v1.UserPhoneChangeReq) (res *v1.UserPhoneChangeRes, err error) {
	number, err := c.runtimeNumber(ctx)
	if err != nil {
		return nil, err
	}
	if err := c.uidentitySvc.ChangeRuntimePhone(ctx, uidentitysvc.ChangePhoneInput{
		Number: number,
		Phone:  req.Phone,
		Code:   req.Code,
	}); err != nil {
		return nil, err
	}
	return &v1.UserMutationRes{}, nil
}
