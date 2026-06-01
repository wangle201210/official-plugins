package uidentity

import (
	"context"

	"lina-plugin-linapro-uidentity-cas/backend/api/uidentity/v1"
)

// RuntimeTokenInfo resolves runtime token user info.
func (c *ControllerV1) RuntimeTokenInfo(ctx context.Context, req *v1.RuntimeTokenInfoReq) (res *v1.RuntimeTokenInfoRes, err error) {
	out, err := c.uidentitySvc.GetUserInfoByRuntimeToken(ctx, req.AccessToken)
	if err != nil {
		return nil, err
	}
	return &v1.RuntimeTokenInfoRes{
		User:  toAPIRuntimeAccount(out.User),
		Users: toAPIRuntimeAccounts(out.Users),
		App:   toAPIRuntimeApplication(out.App),
	}, nil
}
