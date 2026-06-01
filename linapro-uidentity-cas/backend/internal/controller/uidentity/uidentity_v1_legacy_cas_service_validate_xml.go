// This file adapts legacy CAS XML serviceValidate requests to the runtime
// service and writes raw XML before the JSON response middleware wraps output.

package uidentity

import (
	"context"

	"github.com/gogf/gf/v2/frame/g"

	"lina-plugin-linapro-uidentity-cas/backend/api/uidentity/v1"
	uidentitysvc "lina-plugin-linapro-uidentity-cas/backend/internal/service/uidentity"
)

// LegacyCASServiceValidateXML writes old CAS serviceValidate XML.
func (c *ControllerV1) LegacyCASServiceValidateXML(ctx context.Context, req *v1.LegacyCASServiceValidateXMLReq) (res *v1.LegacyCASServiceValidateXMLRes, err error) {
	out, err := c.uidentitySvc.LegacyCASServiceXML(ctx, uidentitysvc.LegacyCASServiceXMLInput{
		Ticket: req.Ticket,
		UserID: req.UserId,
	})
	if err != nil {
		return nil, err
	}
	r := g.RequestFromCtx(ctx)
	r.Response.Header().Set("Content-Type", "application/xml; charset=utf-8")
	r.Response.WriteOver(out.XML)
	r.ExitAll()
	return nil, nil
}
