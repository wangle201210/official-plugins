// admin_v1_list_card.go implements the operator card list handler.

package admin

import (
	"context"

	"lina-plugin-sicau-niu/backend/api/admin/v1"
	cardsvc "lina-plugin-sicau-niu/backend/internal/service/card"
)

// ListCard returns one DB-side paged card list page with batch-assembled
// owning-cattle code and name.
func (c *ControllerV1) ListCard(ctx context.Context, req *v1.ListCardReq) (res *v1.ListCardRes, err error) {
	out, err := c.cardSvc.ListCard(ctx, &cardsvc.ListCardInput{
		Keyword:  req.Keyword,
		Category: req.Category,
		PageNum:  req.PageNum,
		PageSize: req.PageSize,
	})
	if err != nil {
		return nil, err
	}
	items := make([]*v1.CardItem, 0, len(out.List))
	for _, item := range out.List {
		items = append(items, toCardItem(item))
	}
	return &v1.ListCardRes{List: items, Total: out.Total}, nil
}
