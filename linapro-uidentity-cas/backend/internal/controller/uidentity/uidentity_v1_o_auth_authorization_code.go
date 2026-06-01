package uidentity

import (
	"context"

	"lina-plugin-linapro-uidentity-cas/backend/api/uidentity/v1"
	uidentitysvc "lina-plugin-linapro-uidentity-cas/backend/internal/service/uidentity"
)

// OAuthAuthorizationCode validates password login and issues one OAuth code.
func (c *ControllerV1) OAuthAuthorizationCode(ctx context.Context, req *v1.OAuthAuthorizationCodeReq) (res *v1.OAuthAuthorizationCodeRes, err error) {
	out, err := c.uidentitySvc.IssueOAuthAuthorizationCode(ctx, uidentitysvc.OAuthAuthorizationCodeInput{
		ClientID:    req.ClientId,
		RedirectURI: req.RedirectUri,
		Scope:       req.Scope,
		State:       req.State,
		Number:      req.Number,
		Password:    req.Password,
		TtlSeconds:  req.TtlSeconds,
	})
	if err != nil {
		return nil, err
	}
	return &v1.OAuthAuthorizationCodeRes{
		Code:        out.Code,
		RedirectUrl: out.RedirectURL,
		ExpiredAt:   out.ExpiredAt,
		State:       out.State,
	}, nil
}
