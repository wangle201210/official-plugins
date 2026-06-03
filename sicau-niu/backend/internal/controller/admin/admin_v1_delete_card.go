// admin_v1_delete_card.go implements the operator delete-card handler.

package admin

import (
	"context"

	"lina-plugin-sicau-niu/backend/api/admin/v1"
)

// DeleteCard soft-deletes one card, freeing its owning cattle to be bound again.
func (c *ControllerV1) DeleteCard(ctx context.Context, req *v1.DeleteCardReq) (res *v1.DeleteCardRes, err error) {
	err = c.cardSvc.DeleteCard(ctx, req.Id)
	if err != nil {
		return nil, err
	}
	return &v1.DeleteCardRes{}, nil
}
