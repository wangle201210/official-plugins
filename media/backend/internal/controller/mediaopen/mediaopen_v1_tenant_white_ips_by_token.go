// This file implements the public token-based tenant whitelist IP endpoint.

package mediaopen

import (
	"context"

	"lina-plugin-media/backend/api/mediaopen/v1"
	mediasvc "lina-plugin-media/backend/internal/service/media"
)

// TenantWhiteIPsByToken returns enabled tenant whitelist IPs for the token owner.
func (c *ControllerV1) TenantWhiteIPsByToken(ctx context.Context, req *v1.TenantWhiteIPsByTokenReq) (res *v1.TenantWhiteIPsByTokenRes, err error) {
	out, err := c.mediaSvc.ListTenantWhiteIPsByToken(ctx, mediasvc.TenantWhiteIPsByTokenInput{
		Token: req.Token,
	})
	if err != nil {
		return nil, err
	}
	return &v1.TenantWhiteIPsByTokenRes{
		TenantId: out.TenantId,
		Ips:      out.Ips,
	}, nil
}
