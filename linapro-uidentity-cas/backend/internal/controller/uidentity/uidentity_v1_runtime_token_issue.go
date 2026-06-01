package uidentity

import (
	"context"

	"lina-plugin-linapro-uidentity-cas/backend/api/uidentity/v1"
	uidentitysvc "lina-plugin-linapro-uidentity-cas/backend/internal/service/uidentity"
)

// RuntimeTokenIssue issues one legacy runtime access token.
func (c *ControllerV1) RuntimeTokenIssue(ctx context.Context, req *v1.RuntimeTokenIssueReq) (res *v1.RuntimeTokenIssueRes, err error) {
	out, err := c.uidentitySvc.IssueRuntimeToken(ctx, uidentitysvc.RuntimeTokenInput{
		ClientID: req.ClientId,
		Secret:   req.Secret,
		Number:   req.Number,
		Password: req.Password,
	})
	if err != nil {
		return nil, err
	}
	return &v1.RuntimeTokenIssueRes{AccessToken: out.AccessToken, ExpiredAt: out.ExpiredAt}, nil
}
