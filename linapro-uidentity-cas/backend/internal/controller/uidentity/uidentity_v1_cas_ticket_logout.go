package uidentity

import (
	"context"

	"lina-plugin-linapro-uidentity-cas/backend/api/uidentity/v1"
)

// CasTicketLogout deletes one CAS runtime ticket.
func (c *ControllerV1) CasTicketLogout(ctx context.Context, req *v1.CasTicketLogoutReq) (res *v1.CasTicketLogoutRes, err error) {
	if err := c.uidentitySvc.DeleteTicket(ctx, req.Ticket); err != nil {
		return nil, err
	}
	return &v1.CasTicketLogoutRes{}, nil
}
