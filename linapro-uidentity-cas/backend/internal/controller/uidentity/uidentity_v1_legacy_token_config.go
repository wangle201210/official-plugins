// This file maps plugin-scoped legacy runtime token static configuration into
// the old admin response shape.

package uidentity

import (
	"context"

	"lina-plugin-linapro-uidentity-cas/backend/api/uidentity/v1"
)

// LegacyTokenConfig returns old admin runtime token static configuration metadata.
func (c *ControllerV1) LegacyTokenConfig(ctx context.Context, req *v1.LegacyTokenConfigReq) (res *v1.LegacyTokenConfigRes, err error) {
	out, err := c.uidentitySvc.LegacyTokenConfig(ctx)
	if err != nil {
		return nil, err
	}
	return &v1.LegacyTokenConfigRes{
		GetAddr:   out.GetAddr,
		CheckAddr: out.CheckAddr,
		TokenDocs: out.TokenDocs,
		CasDocs:   out.CasDocs,
	}, nil
}
