package uidentity

import (
	"context"

	"lina-plugin-linapro-uidentity-cas/backend/api/uidentity/v1"
	uidentitysvc "lina-plugin-linapro-uidentity-cas/backend/internal/service/uidentity"
)

// CasServiceTicket issues the old ssologin access token.
func (c *ControllerV1) CasServiceTicket(ctx context.Context, req *v1.CasServiceTicketReq) (res *v1.CasServiceTicketRes, err error) {
	out, err := c.uidentitySvc.IssueRuntimeToken(ctx, uidentitysvc.RuntimeTokenInput{
		ClientID: req.ClientId,
		Secret:   req.Secret,
		Number:   req.Number,
		Password: req.Password,
	})
	if err != nil {
		return nil, err
	}
	return &v1.CasServiceTicketRes{AccessToken: out.AccessToken}, nil
}
