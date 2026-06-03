// admin_v1_update_card.go implements the operator update-card handler.

package admin

import (
	"context"

	"lina-plugin-sicau-niu/backend/api/admin/v1"
	cardsvc "lina-plugin-sicau-niu/backend/internal/service/card"
)

// UpdateCard updates one card.
func (c *ControllerV1) UpdateCard(ctx context.Context, req *v1.UpdateCardReq) (res *v1.UpdateCardRes, err error) {
	err = c.cardSvc.UpdateCard(ctx, req.Id, &cardsvc.CardMutateInput{
		NiuId:     req.NiuId,
		Category:  req.Category,
		Title:     req.Title,
		Content:   req.Content,
		ImagePath: req.ImagePath,
	})
	if err != nil {
		return nil, err
	}
	return &v1.UpdateCardRes{}, nil
}
