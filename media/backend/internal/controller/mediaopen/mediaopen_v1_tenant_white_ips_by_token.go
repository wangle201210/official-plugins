// This file implements the public token-based tenant whitelist IP endpoint.

package mediaopen

import (
	"context"

	"github.com/gogf/gf/v2/frame/g"

	"lina-plugin-media/backend/api/mediaopen/v1"
	mediasvc "lina-plugin-media/backend/internal/service/media"
)

// TenantWhiteIPsByToken returns enabled tenant whitelist IPs for the token owner.
func (c *ControllerV1) TenantWhiteIPsByToken(ctx context.Context, req *v1.TenantWhiteIPsByTokenReq) (res *v1.TenantWhiteIPsByTokenRes, err error) {
	ips, err := c.mediaSvc.ListTenantWhiteIPsByToken(ctx, mediasvc.TenantWhiteIPsByTokenInput{
		Token: req.Token,
	})
	if err != nil {
		return nil, err
	}
	if request := g.RequestFromCtx(ctx); request != nil {
		request.Response.WriteJson(ips)
		return nil, nil
	}
	out := v1.TenantWhiteIPsByTokenRes(ips)
	return &out, nil
}
