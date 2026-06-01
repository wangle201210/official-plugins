package uidentity

import (
	"context"

	"lina-plugin-linapro-uidentity-cas/backend/api/uidentity/v1"
)

// OAuthAccessTokenInfo resolves OAuth token user information.
func (c *ControllerV1) OAuthAccessTokenInfo(ctx context.Context, req *v1.OAuthAccessTokenInfoReq) (res *v1.OAuthAccessTokenInfoRes, err error) {
	out, err := c.uidentitySvc.GetOAuthAccessTokenInfo(ctx, req.AccessToken)
	if err != nil {
		return nil, err
	}
	return &v1.OAuthAccessTokenInfoRes{
		User:  toAPIRuntimeAccount(out.User),
		App:   toAPIRuntimeApplication(out.App),
		Scope: out.Scope,
	}, nil
}
