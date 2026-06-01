package uidentity

import (
	"context"

	"lina-plugin-linapro-uidentity-cas/backend/api/uidentity/v1"
	uidentitysvc "lina-plugin-linapro-uidentity-cas/backend/internal/service/uidentity"
)

// UserLoginLogs returns paged CAS logs for one runtime account.
func (c *ControllerV1) UserLoginLogs(ctx context.Context, req *v1.UserLoginLogsReq) (res *v1.UserLoginLogsRes, err error) {
	out, err := c.uidentitySvc.ListRuntimeUserLoginLogs(ctx, uidentitysvc.UserLogListInput{
		Number:   req.Number,
		PageNum:  req.PageNum,
		PageSize: req.PageSize,
	})
	if err != nil {
		return nil, err
	}
	return &v1.UserLoginLogsRes{List: toAPIRecords(out.List), Total: out.Total}, nil
}
