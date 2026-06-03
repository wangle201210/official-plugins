// admin_v1_get_niu.go implements the operator get-cattle-detail handler.

package admin

import (
	"context"

	"lina-plugin-sicau-niu/backend/api/admin/v1"
)

// GetNiu returns one cattle detail with its batch-assembled relations.
func (c *ControllerV1) GetNiu(ctx context.Context, req *v1.GetNiuReq) (res *v1.GetNiuRes, err error) {
	item, err := c.cattleSvc.GetNiu(ctx, req.Id)
	if err != nil {
		return nil, err
	}
	return &v1.GetNiuRes{NiuItem: toNiuItem(item)}, nil
}
