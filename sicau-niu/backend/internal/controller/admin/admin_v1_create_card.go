// admin_v1_create_card.go implements the operator create-card handler.

package admin

import (
	"context"

	"lina-plugin-sicau-niu/backend/api/admin/v1"
	cardsvc "lina-plugin-sicau-niu/backend/internal/service/card"
)

// CreateCard creates one card bound to a cattle.
func (c *ControllerV1) CreateCard(ctx context.Context, req *v1.CreateCardReq) (res *v1.CreateCardRes, err error) {
	id, err := c.cardSvc.CreateCard(ctx, &cardsvc.CardMutateInput{
		NiuId:     req.NiuId,
		Category:  req.Category,
		Title:     req.Title,
		Content:   req.Content,
		ImagePath: req.ImagePath,
	})
	if err != nil {
		return nil, err
	}
	return &v1.CreateCardRes{Id: id}, nil
}
