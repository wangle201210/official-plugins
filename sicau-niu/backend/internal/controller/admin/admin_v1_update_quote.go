// admin_v1_update_quote.go implements the operator update-quote handler.

package admin

import (
	"context"

	"lina-plugin-sicau-niu/backend/api/admin/v1"
	cardsvc "lina-plugin-sicau-niu/backend/internal/service/card"
)

// UpdateQuote updates one quote.
func (c *ControllerV1) UpdateQuote(ctx context.Context, req *v1.UpdateQuoteReq) (res *v1.UpdateQuoteRes, err error) {
	err = c.cardSvc.UpdateQuote(ctx, req.Id, &cardsvc.QuoteMutateInput{
		Content: req.Content,
		Enabled: req.Enabled,
	})
	if err != nil {
		return nil, err
	}
	return &v1.UpdateQuoteRes{}, nil
}
