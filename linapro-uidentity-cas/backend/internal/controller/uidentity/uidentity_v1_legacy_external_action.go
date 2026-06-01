// This file adapts legacy external dependency action requests to the plugin
// service boundary without constructing external executors in the controller.

package uidentity

import (
	"context"

	"lina-plugin-linapro-uidentity-cas/backend/api/uidentity/v1"
	uidentitysvc "lina-plugin-linapro-uidentity-cas/backend/internal/service/uidentity"
)

// LegacyExternalAction exposes the legacy external dependency boundary.
func (c *ControllerV1) LegacyExternalAction(ctx context.Context, req *v1.LegacyExternalActionReq) (res *v1.LegacyExternalActionRes, err error) {
	out, err := c.uidentitySvc.RunExternalAction(ctx, uidentitysvc.LegacyExternalActionInput{
		Type:   req.Type,
		Target: req.Target,
	})
	if err != nil {
		return nil, err
	}
	return &v1.LegacyExternalActionRes{
		Type:    out.Type,
		Target:  out.Target,
		Success: out.Success,
	}, nil
}
