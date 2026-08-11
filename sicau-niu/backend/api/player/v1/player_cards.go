// player_cards.go defines the request and response DTOs for the current player's
// personal card collection (图鉴): the bounded card catalog plus the current
// player's ownership state, optionally filtered by category. Locked entries keep
// their category for progress calculation while withholding card-face content.

package v1

import "github.com/gogf/gf/v2/frame/g"

// CollectionReq is the request for the current player's personal card collection.
type CollectionReq struct {
	g.Meta   `path:"/plugins/sicau-niu/player/cards" method:"get" tags:"寻牛小程序" summary:"查询玩家个人图鉴" dc:"Return the bounded card catalog with the authenticated player's ownership state. Locked entries retain cattle reference and category but omit title, content and image so the mini-program can render silhouettes and five-category progress without exposing card-face content. The optional category filter narrows the catalog. Data is assembled with bounded batch queries to avoid N+1. Requires a valid player token."`
	Category string `json:"category" dc:"Optional card category filter: person=人物, event=事件, research=科研, college=院系, spirit=精神; empty returns the complete bounded catalog" eg:"spirit"`
}

// CollectionRes is the response for the current player's personal card collection.
type CollectionRes struct {
	List []*CollectionCardItem `json:"list" dc:"Complete bounded card catalog ordered by card ID, with per-player ownership state" eg:"[]"`
}

// CollectionCardItem defines one collected card with its owning cattle reference.
type CollectionCardItem struct {
	NiuId     int64  `json:"niuId" dc:"Owning cattle ID" eg:"1"`
	NiuCode   string `json:"niuCode" dc:"Owning cattle serial code" eg:"NIU-001"`
	Category  string `json:"category" dc:"Card category: person=人物, event=事件, research=科研, college=院系, spirit=精神" eg:"spirit"`
	Owned     bool   `json:"owned" dc:"Whether the current player owns this card" eg:"true"`
	Title     string `json:"title" dc:"Card title; empty while the card is locked" eg:"川农大精神"`
	Content   string `json:"content" dc:"Card content text; empty while the card is locked" eg:"爱国敬业、艰苦奋斗、团结拼搏、求实创新"`
	ImagePath string `json:"imagePath" dc:"Card image storage path; empty while locked or when no image is configured" eg:"sicau-niu/card/1.jpg"`
}
