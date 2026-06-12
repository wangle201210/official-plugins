// admin_quote_list.go defines the request and response DTOs for the operator
// quote list query.

package v1

import "github.com/gogf/gf/v2/frame/g"

// ListQuoteReq is the request for the operator quote list query.
type ListQuoteReq struct {
	g.Meta   `path:"/plugins/sicau-niu/admin/quotes" method:"get" tags:"Sicau Niu Admin" summary:"查询金句列表" dc:"List school-history quotes with optional content fuzzy filtering, ordered by ID descending, with DB-side pagination. Protected by host unified permission check." permission:"sicau-niu:quote:list"`
	Keyword  string `json:"keyword" dc:"Fuzzy filter by quote content; lists all when omitted" eg:"自强不息"`
	PageNum  int    `json:"pageNum" dc:"Page number; defaults to 1 when omitted or non-positive" eg:"1"`
	PageSize int    `json:"pageSize" dc:"Items per page; defaults to 10 and is capped at 100" eg:"10"`
}

// ListQuoteRes is the response for the operator quote list query.
type ListQuoteRes struct {
	List  []*QuoteItem `json:"list" dc:"Quotes on the current page" eg:"[]"`
	Total int          `json:"total" dc:"Total number of matching quotes" eg:"1"`
}

// QuoteItem defines one quote row for the operator console.
type QuoteItem struct {
	Id        int64  `json:"id" dc:"Quote ID" eg:"1"`
	Content   string `json:"content" dc:"Quote text" eg:"追求真理 造福社会 自强不息"`
	Enabled   int    `json:"enabled" dc:"Whether the quote participates in random playback: 1=启用, 0=禁用" eg:"1"`
	CreatedAt *int64 `json:"createdAt" dc:"Creation time as Unix timestamp in milliseconds" eg:"1776333600000"`
	UpdatedAt *int64 `json:"updatedAt" dc:"Update time as Unix timestamp in milliseconds" eg:"1776333900000"`
}
