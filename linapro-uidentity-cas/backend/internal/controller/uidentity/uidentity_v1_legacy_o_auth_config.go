// This file maps plugin-scoped legacy OAuth static configuration into the old
// admin response shape without coupling to legacy page routes.

package uidentity

import (
	"context"

	"lina-plugin-linapro-uidentity-cas/backend/api/uidentity/v1"
)

// LegacyOAuthConfig returns old admin OAuth static configuration metadata.
func (c *ControllerV1) LegacyOAuthConfig(ctx context.Context, req *v1.LegacyOAuthConfigReq) (res *v1.LegacyOAuthConfigRes, err error) {
	out, err := c.uidentitySvc.LegacyOAuthConfig(ctx)
	if err != nil {
		return nil, err
	}
	return &v1.LegacyOAuthConfigRes{
		Authorization: out.Authorization,
		GetTokenAddr:  out.GetTokenAddr,
		UserInfoAddr:  out.UserInfoAddr,
		LogoutAddr:    out.LogoutAddr,
		PingAddr:      out.PingAddr,
		Docs:          out.Docs,
	}, nil
}
