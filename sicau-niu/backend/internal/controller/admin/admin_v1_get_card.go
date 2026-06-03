// admin_v1_get_card.go implements the operator get-card-detail handler.

package admin

import (
	"context"

	"lina-plugin-sicau-niu/backend/api/admin/v1"
)

// GetCard returns one card detail with its batch-assembled owning-cattle info.
func (c *ControllerV1) GetCard(ctx context.Context, req *v1.GetCardReq) (res *v1.GetCardRes, err error) {
	item, err := c.cardSvc.GetCard(ctx, req.Id)
	if err != nil {
		return nil, err
	}
	return &v1.GetCardRes{CardItem: toCardItem(item)}, nil
}
