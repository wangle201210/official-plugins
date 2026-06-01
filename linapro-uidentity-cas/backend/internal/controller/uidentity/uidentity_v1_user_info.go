package uidentity

import (
	"context"

	"lina-plugin-linapro-uidentity-cas/backend/api/uidentity/v1"
)

// UserInfo returns one runtime account projection.
func (c *ControllerV1) UserInfo(ctx context.Context, req *v1.UserInfoReq) (res *v1.UserInfoRes, err error) {
	out, err := c.uidentitySvc.GetRuntimeUserInfo(ctx, req.Number)
	if err != nil {
		return nil, err
	}
	return &v1.UserInfoRes{User: toAPIRuntimeAccount(out)}, nil
}
