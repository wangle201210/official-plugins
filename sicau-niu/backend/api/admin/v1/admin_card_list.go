// admin_card_list.go defines the request and response DTOs for the operator card
// list query, including the batch-assembled owning-cattle code and name.

package v1

import "github.com/gogf/gf/v2/frame/g"

// ListCardReq is the request for the operator card list query.
type ListCardReq struct {
	g.Meta   `path:"/plugins/sicau-niu/admin/cards" method:"get" tags:"Sicau Niu Admin" summary:"查询卡片列表" dc:"List cards with optional title fuzzy filtering and category filtering, ordered by ID descending, with DB-side pagination. The owning-cattle code and name are batch-assembled to avoid N+1 queries. Protected by host unified permission check." permission:"sicau-niu:card:list"`
	Keyword  string `json:"keyword" dc:"Fuzzy filter by card title; lists all when omitted" eg:"校训"`
	Category string `json:"category" dc:"Filter by category: person=人物, event=事件, research=科研, college=院系, spirit=精神; lists all when omitted" eg:"person"`
	PageNum  int    `json:"pageNum" dc:"Page number; defaults to 1 when omitted or non-positive" eg:"1"`
	PageSize int    `json:"pageSize" dc:"Items per page; defaults to 10 and is capped at 100" eg:"10"`
}

// ListCardRes is the response for the operator card list query.
type ListCardRes struct {
	List  []*CardItem `json:"list" dc:"Cards on the current page" eg:"[]"`
	Total int         `json:"total" dc:"Total number of matching cards" eg:"1"`
}

// CardItem defines one card row for the operator console.
type CardItem struct {
	Id        int64  `json:"id" dc:"Card ID" eg:"1"`
	NiuId     int64  `json:"niuId" dc:"Owning cattle ID; one card per cattle" eg:"1"`
	NiuCode   string `json:"niuCode" dc:"Owning cattle code, batch-assembled; empty when the cattle is missing" eg:"N001"`
	NiuName   string `json:"niuName" dc:"Owning cattle name, batch-assembled; empty when unset or cattle missing" eg:"信息工程学院牛"`
	Category  string `json:"category" dc:"Card category: person=人物, event=事件, research=科研, college=院系, spirit=精神" eg:"person"`
	Title     string `json:"title" dc:"Card title" eg:"川农大校训"`
	Content   string `json:"content" dc:"Card content text" eg:"追求真理 造福社会 自强不息"`
	ImagePath string `json:"imagePath" dc:"Card image storage path uploaded via host file management; empty when none" eg:"/uploads/2026/06/card.png"`
	CreatedAt *int64 `json:"createdAt" dc:"Creation time as Unix timestamp in milliseconds" eg:"1776333600000"`
	UpdatedAt *int64 `json:"updatedAt" dc:"Update time as Unix timestamp in milliseconds" eg:"1776333900000"`
}
