package uidentity

import (
	"context"

	"lina-plugin-linapro-uidentity-cas/backend/api/uidentity/v1"
	uidentitysvc "lina-plugin-linapro-uidentity-cas/backend/internal/service/uidentity"
)

// OAuthAccessToken exchanges one authorization code for OAuth tokens.
func (c *ControllerV1) OAuthAccessToken(ctx context.Context, req *v1.OAuthAccessTokenReq) (res *v1.OAuthAccessTokenRes, err error) {
	out, err := c.uidentitySvc.ExchangeOAuthAuthorizationCode(ctx, uidentitysvc.OAuthTokenExchangeInput{
		GrantType:    req.GrantType,
		ClientID:     req.ClientId,
		ClientSecret: req.ClientSecret,
		Code:         req.Code,
		RedirectURI:  req.RedirectUri,
		TtlSeconds:   req.TtlSeconds,
	})
	if err != nil {
		return nil, err
	}
	return &v1.OAuthAccessTokenRes{
		AccessToken:  out.AccessToken,
		RefreshToken: out.RefreshToken,
		TokenType:    out.TokenType,
		ExpiresIn:    out.ExpiresIn,
		ExpiredAt:    out.ExpiredAt,
		Scope:        out.Scope,
	}, nil
}
