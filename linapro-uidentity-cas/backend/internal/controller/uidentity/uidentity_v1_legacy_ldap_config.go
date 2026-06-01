// This file maps plugin-scoped legacy LDAP static configuration into the old
// admin response shape while preserving the default unsupported executor state.

package uidentity

import (
	"context"

	"lina-plugin-linapro-uidentity-cas/backend/api/uidentity/v1"
)

// LegacyLDAPConfig returns old admin LDAP static configuration metadata.
func (c *ControllerV1) LegacyLDAPConfig(ctx context.Context, req *v1.LegacyLDAPConfigReq) (res *v1.LegacyLDAPConfigRes, err error) {
	out, err := c.uidentitySvc.LegacyLDAPConfig(ctx)
	if err != nil {
		return nil, err
	}
	return &v1.LegacyLDAPConfigRes{
		Version: out.Version,
		Addr:    out.Addr,
		Docs:    out.Docs,
	}, nil
}
