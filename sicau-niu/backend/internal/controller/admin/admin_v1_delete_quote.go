// admin_v1_delete_quote.go implements the operator delete-quote handler.

package admin

import (
	"context"

	"lina-plugin-sicau-niu/backend/api/admin/v1"
)

// DeleteQuote soft-deletes one quote.
func (c *ControllerV1) DeleteQuote(ctx context.Context, req *v1.DeleteQuoteReq) (res *v1.DeleteQuoteRes, err error) {
	err = c.cardSvc.DeleteQuote(ctx, req.Id)
	if err != nil {
		return nil, err
	}
	return &v1.DeleteQuoteRes{}, nil
}
