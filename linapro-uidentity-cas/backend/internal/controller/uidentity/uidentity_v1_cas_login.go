package uidentity

import (
	"context"

	"lina-plugin-linapro-uidentity-cas/backend/api/uidentity/v1"
	uidentitysvc "lina-plugin-linapro-uidentity-cas/backend/internal/service/uidentity"
)

// CasLogin validates a CAS ticket and records a login log.
func (c *ControllerV1) CasLogin(ctx context.Context, req *v1.CasLoginReq) (res *v1.CasLoginRes, err error) {
	out, err := c.uidentitySvc.LoginByCASTicket(ctx, uidentitysvc.CASLoginInput{
		Ticket: req.Ticket,
		UserID: req.UserId,
		AppID:  req.AppId,
	})
	if err != nil {
		return nil, err
	}
	return &v1.CasLoginRes{Number: out.Number, AccountId: out.AccountID, AppId: out.AppID}, nil
}
