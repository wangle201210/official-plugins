// admin_card_update.go defines the request and response DTOs for updating one
// card.

package v1

import "github.com/gogf/gf/v2/frame/g"

// UpdateCardReq is the request for updating one card.
type UpdateCardReq struct {
	g.Meta    `path:"/plugins/sicau-niu/admin/cards/{id}" method:"put" tags:"Sicau Niu Admin" summary:"Update card" dc:"Update a card's owning cattle, category, title, content and image path. Re-binding to another cattle keeps the one-card-per-cattle constraint and requires the target cattle to exist. Protected by host unified permission check." permission:"sicau-niu:card:update"`
	Id        int64  `json:"id" v:"required|min:1" dc:"Card ID from the path" eg:"1"`
	NiuId     int64  `json:"niuId" v:"required|min:1" dc:"Owning cattle ID; must exist and must not already have another card" eg:"1"`
	Category  string `json:"category" v:"required" dc:"Card category: person=人物, event=事件, research=科研, college=院系, spirit=精神" eg:"person"`
	Title     string `json:"title" v:"required|length:1,128" dc:"Card title" eg:"川农大校训"`
	Content   string `json:"content" dc:"Card content text" eg:"追求真理 造福社会 自强不息"`
	ImagePath string `json:"imagePath" dc:"Card image storage path obtained from the host file upload; empty clears it" eg:"/uploads/2026/06/card.png"`
}

// UpdateCardRes is the response for updating one card. It is intentionally empty;
// success is conveyed by the absence of an error.
type UpdateCardRes struct{}
