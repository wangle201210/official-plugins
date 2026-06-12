// admin_quote_create.go defines the request and response DTOs for creating one
// quote.

package v1

import "github.com/gogf/gf/v2/frame/g"

// CreateQuoteReq is the request for creating one quote.
type CreateQuoteReq struct {
	g.Meta  `path:"/plugins/sicau-niu/admin/quotes" method:"post" tags:"Sicau Niu Admin" summary:"新增金句" dc:"Create a non-empty school-history quote for the random-playback pool. Enabled defaults to 1 when omitted. Protected by host unified permission check." permission:"sicau-niu:quote:create"`
	Content string `json:"content" v:"required|length:1,512" dc:"Quote text; must be non-empty" eg:"追求真理 造福社会 自强不息"`
	Enabled *int   `json:"enabled" dc:"Whether the quote participates in random playback: 1=启用, 0=禁用; defaults to 1 when omitted" eg:"1"`
}

// CreateQuoteRes is the response for creating one quote.
type CreateQuoteRes struct {
	Id int64 `json:"id" dc:"The newly created quote ID" eg:"1"`
}
