// admin_quote_delete.go defines the request and response DTOs for deleting one
// quote.

package v1

import "github.com/gogf/gf/v2/frame/g"

// DeleteQuoteReq is the request for deleting one quote.
type DeleteQuoteReq struct {
	g.Meta `path:"/plugins/sicau-niu/admin/quotes/{id}" method:"delete" tags:"Sicau Niu Admin" summary:"Delete quote" dc:"Soft-delete a quote. Recoverable. Protected by host unified permission check." permission:"sicau-niu:quote:delete"`
	Id     int64 `json:"id" v:"required|min:1" dc:"Quote ID from the path" eg:"1"`
}

// DeleteQuoteRes is the response for deleting one quote. It is intentionally
// empty; success is conveyed by the absence of an error.
type DeleteQuoteRes struct{}
