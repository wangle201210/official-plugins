package uidentity

import (
	"context"

	"lina-plugin-linapro-uidentity-cas/backend/api/uidentity/v1"
)

// UserQQChange changes one runtime account QQ number.
func (c *ControllerV1) UserQQChange(ctx context.Context, req *v1.UserQQChangeReq) (res *v1.UserQQChangeRes, err error) {
	if err := c.uidentitySvc.ChangeRuntimeQQ(ctx, req.Number, req.Qq); err != nil {
		return nil, err
	}
	return &v1.UserMutationRes{}, nil
}
