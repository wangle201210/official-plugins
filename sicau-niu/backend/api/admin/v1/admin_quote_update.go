// admin_quote_update.go defines the request and response DTOs for updating one
// quote.

package v1

import "github.com/gogf/gf/v2/frame/g"

// UpdateQuoteReq is the request for updating one quote.
type UpdateQuoteReq struct {
	g.Meta  `path:"/plugins/sicau-niu/admin/quotes/{id}" method:"put" tags:"Sicau Niu Admin" summary:"修改金句" dc:"Update a quote's content and enabled flag. The content must stay non-empty. Protected by host unified permission check." permission:"sicau-niu:quote:update"`
	Id      int64  `json:"id" v:"required|min:1" dc:"Quote ID from the path" eg:"1"`
	Content string `json:"content" v:"required|length:1,512" dc:"Quote text; must be non-empty" eg:"追求真理 造福社会 自强不息"`
	Enabled *int   `json:"enabled" dc:"Whether the quote participates in random playback: 1=启用, 0=禁用; unchanged when omitted" eg:"1"`
}

// UpdateQuoteRes is the response for updating one quote. It is intentionally
// empty; success is conveyed by the absence of an error.
type UpdateQuoteRes struct{}
