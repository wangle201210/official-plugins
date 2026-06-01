package uidentity

import (
	"context"

	"lina-plugin-linapro-uidentity-cas/backend/api/uidentity/v1"
	uidentitysvc "lina-plugin-linapro-uidentity-cas/backend/internal/service/uidentity"
)

// CasServiceValidate consumes and validates a CAS service ticket.
func (c *ControllerV1) CasServiceValidate(ctx context.Context, req *v1.CasServiceValidateReq) (res *v1.CasServiceValidateRes, err error) {
	out, err := c.uidentitySvc.ValidateServiceTicket(ctx, uidentitysvc.ServiceValidateInput{
		Ticket: req.Ticket,
		UserID: req.UserId,
	})
	if err != nil {
		return nil, err
	}
	return &v1.CasServiceValidateRes{
		Ticket:  out.Ticket,
		User:    toAPIRuntimeAccount(out.User),
		App:     toAPIRuntimeApplication(out.App),
		Success: out.Success,
	}, nil
}
