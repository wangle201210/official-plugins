package uidentity

import (
	"context"

	"lina-plugin-linapro-uidentity-cas/backend/api/uidentity/v1"
	uidentitysvc "lina-plugin-linapro-uidentity-cas/backend/internal/service/uidentity"
)

// OAuthIssue issues an OAuth token record and log.
func (c *ControllerV1) OAuthIssue(ctx context.Context, req *v1.OAuthIssueReq) (res *v1.OAuthIssueRes, err error) {
	out, err := c.uidentitySvc.IssueOAuthToken(ctx, uidentitysvc.OAuthIssueInput{
		AccountID:   req.AccountId,
		AppID:       req.AppId,
		RedirectURI: req.RedirectUri,
		Scope:       req.Scope,
		TtlSeconds:  req.TtlSeconds,
	})
	if err != nil {
		return nil, err
	}
	return &v1.OAuthIssueRes{
		Code:      out.Code,
		Access:    out.Access,
		Refresh:   out.Refresh,
		ExpiredAt: out.ExpiredAt,
	}, nil
}
