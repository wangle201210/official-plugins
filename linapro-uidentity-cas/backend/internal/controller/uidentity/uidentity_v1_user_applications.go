package uidentity

import (
	"context"

	"lina-plugin-linapro-uidentity-cas/backend/api/uidentity/v1"
	uidentitysvc "lina-plugin-linapro-uidentity-cas/backend/internal/service/uidentity"
)

// UserApplications returns accessible applications for one runtime account.
func (c *ControllerV1) UserApplications(ctx context.Context, req *v1.UserApplicationsReq) (res *v1.UserApplicationsRes, err error) {
	out, err := c.uidentitySvc.ListRuntimeApplications(ctx, uidentitysvc.UserApplicationListInput{
		Number:   req.Number,
		PageNum:  req.PageNum,
		PageSize: req.PageSize,
	})
	if err != nil {
		return nil, err
	}
	return &v1.UserApplicationsRes{List: toAPIRuntimeApplications(out.List), Total: out.Total}, nil
}
