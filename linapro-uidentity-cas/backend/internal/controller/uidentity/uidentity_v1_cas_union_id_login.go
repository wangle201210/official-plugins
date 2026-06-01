package uidentity

import (
	"context"

	"lina-plugin-linapro-uidentity-cas/backend/api/uidentity/v1"
	uidentitysvc "lina-plugin-linapro-uidentity-cas/backend/internal/service/uidentity"
)

// CasUnionIDLogin validates union ID login and returns CAS tickets.
func (c *ControllerV1) CasUnionIDLogin(ctx context.Context, req *v1.CasUnionIDLoginReq) (res *v1.CasUnionIDLoginRes, err error) {
	out, err := c.uidentitySvc.LoginByUnionID(ctx, uidentitysvc.UnionIDLoginInput{
		ClientID: req.ClientId,
		UnionID:  req.UnionId,
	})
	if err != nil {
		return nil, err
	}
	return toAPIRuntimeLogin(out), nil
}
