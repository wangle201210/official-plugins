// admin_v1_revoke_activation.go implements activation repair.
package admin

import (
	"context"

	"lina-plugin-sicau-niu/backend/api/admin/v1"
)

func (c *ControllerV1) RevokeActivation(ctx context.Context, req *v1.RevokeActivationReq) (*v1.RevokeActivationRes, error) {
	if err := c.activationSvc.Revoke(ctx, req.Id); err != nil {
		return nil, err
	}
	return &v1.RevokeActivationRes{}, nil
}
