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
	out, err := c.mediaSvc.ListTenantWhiteIPsByToken(ctx, mediasvc.TenantWhiteIPsByTokenInput{
		Token: req.Token,
	})
	if err != nil {
		return nil, err
	}
	res = &v1.TenantWhiteIPsByTokenRes{
		TenantId: out.TenantId,
		Ips:      out.Ips,
	}
	if request := g.RequestFromCtx(ctx); request != nil {
		request.Response.WriteJson(res)
		return nil, nil
	}
	return res, nil
}
