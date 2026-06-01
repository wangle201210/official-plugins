// This file exposes the legacy health-check endpoint through the generated
// controller layer and delegates all response data to the UIdentity service.

package uidentity

import (
	"context"

	"lina-plugin-linapro-uidentity-cas/backend/api/uidentity/v1"
)

// LegacyHealth returns a lightweight health check response.
func (c *ControllerV1) LegacyHealth(ctx context.Context, req *v1.LegacyHealthReq) (res *v1.LegacyHealthRes, err error) {
	out, err := c.uidentitySvc.Health(ctx)
	if err != nil {
		return nil, err
	}
	return &v1.LegacyHealthRes{Status: out.Status}, nil
}
