// admin_v1_create_quote.go implements the operator create-quote handler.

package admin

import (
	"context"

	"lina-plugin-sicau-niu/backend/api/admin/v1"
	cardsvc "lina-plugin-sicau-niu/backend/internal/service/card"
)

// CreateQuote creates one quote.
func (c *ControllerV1) CreateQuote(ctx context.Context, req *v1.CreateQuoteReq) (res *v1.CreateQuoteRes, err error) {
	id, err := c.cardSvc.CreateQuote(ctx, &cardsvc.QuoteMutateInput{
		Content: req.Content,
		Enabled: req.Enabled,
	})
	if err != nil {
		return nil, err
	}
	return &v1.CreateQuoteRes{Id: id}, nil
}
