// This file maps plugin-scoped legacy CAS static configuration into the old
// admin response shape without reading host-global configuration.

package uidentity

import (
	"context"

	"lina-plugin-linapro-uidentity-cas/backend/api/uidentity/v1"
)

// LegacyCASConfig returns old admin CAS static configuration metadata.
func (c *ControllerV1) LegacyCASConfig(ctx context.Context, req *v1.LegacyCASConfigReq) (res *v1.LegacyCASConfigRes, err error) {
	out, err := c.uidentitySvc.LegacyCASConfig(ctx)
	if err != nil {
		return nil, err
	}
	return &v1.LegacyCASConfigRes{
		LoginAddr:  out.LoginAddr,
		LogoutAddr: out.LogoutAddr,
		RestAddr:   out.RestAddr,
		Docs:       out.Docs,
	}, nil
}
