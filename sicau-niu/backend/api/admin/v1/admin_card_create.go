// admin_card_create.go defines the request and response DTOs for creating one
// card bound to a cattle.

package v1

import "github.com/gogf/gf/v2/frame/g"

// CreateCardReq is the request for creating one card.
type CreateCardReq struct {
	g.Meta    `path:"/plugins/sicau-niu/admin/cards" method:"post" tags:"Sicau Niu Admin" summary:"Create card" dc:"Create a card bound to an existing cattle under the one-card-per-cattle constraint. The category must be a valid enum. The image path comes from the host file upload. Protected by host unified permission check." permission:"sicau-niu:card:create"`
	NiuId     int64  `json:"niuId" v:"required|min:1" dc:"Owning cattle ID; must exist and must not already have a card" eg:"1"`
	Category  string `json:"category" v:"required" dc:"Card category: person=人物, event=事件, research=科研, college=院系, spirit=精神" eg:"person"`
	Title     string `json:"title" v:"required|length:1,128" dc:"Card title" eg:"川农大校训"`
	Content   string `json:"content" dc:"Card content text" eg:"追求真理 造福社会 自强不息"`
	ImagePath string `json:"imagePath" dc:"Card image storage path obtained from the host file upload; empty when none" eg:"/uploads/2026/06/card.png"`
}

// CreateCardRes is the response for creating one card.
type CreateCardRes struct {
	Id int64 `json:"id" dc:"The newly created card ID" eg:"1"`
}
