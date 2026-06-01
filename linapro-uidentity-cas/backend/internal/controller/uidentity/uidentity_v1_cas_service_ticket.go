package uidentity

import (
	"context"

	"lina-plugin-linapro-uidentity-cas/backend/api/uidentity/v1"
	uidentitysvc "lina-plugin-linapro-uidentity-cas/backend/internal/service/uidentity"
)

// CasServiceTicket issues a CAS service ticket from a TGT.
func (c *ControllerV1) CasServiceTicket(ctx context.Context, req *v1.CasServiceTicketReq) (res *v1.CasServiceTicketRes, err error) {
	out, err := c.uidentitySvc.IssueServiceTicketFromTGT(ctx, uidentitysvc.ServiceTicketInput{
		ClientID:  req.ClientId,
		TGT:       req.Tgt,
		AccountID: req.AccountId,
	})
	if err != nil {
		return nil, err
	}
	return &v1.CasServiceTicketRes{St: out.ST, CallbackUrl: out.CallbackURL}, nil
}
