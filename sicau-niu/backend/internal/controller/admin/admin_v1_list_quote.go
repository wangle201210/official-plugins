// admin_v1_list_quote.go implements the operator quote list handler.

package admin

import (
	"context"

	"lina-plugin-sicau-niu/backend/api/admin/v1"
	cardsvc "lina-plugin-sicau-niu/backend/internal/service/card"
)

// ListQuote returns one DB-side paged quote list page.
func (c *ControllerV1) ListQuote(ctx context.Context, req *v1.ListQuoteReq) (res *v1.ListQuoteRes, err error) {
	out, err := c.cardSvc.ListQuote(ctx, &cardsvc.ListQuoteInput{
		Keyword:  req.Keyword,
		PageNum:  req.PageNum,
		PageSize: req.PageSize,
	})
	if err != nil {
		return nil, err
	}
	items := make([]*v1.QuoteItem, 0, len(out.List))
	for _, item := range out.List {
		items = append(items, &v1.QuoteItem{
			Id:        item.Id,
			Content:   item.Content,
			Enabled:   item.Enabled,
			CreatedAt: item.CreatedAt,
			UpdatedAt: item.UpdatedAt,
		})
	}
	return &v1.ListQuoteRes{List: items, Total: out.Total}, nil
}
