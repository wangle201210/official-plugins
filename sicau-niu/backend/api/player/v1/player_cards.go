// player_cards.go defines the request and response DTOs for the current player's
// personal card collection (图鉴): the main cards of the cattle the player has
// activated, optionally filtered by category. The collection is derived from the
// player's own activation records and is isolated to the current player.

package v1

import "github.com/gogf/gf/v2/frame/g"

// CollectionReq is the request for the current player's personal card collection.
type CollectionReq struct {
	g.Meta   `path:"/plugins/sicau-niu/player/cards" method:"get" tags:"Sicau Niu Player" summary:"List the player personal card collection" dc:"Return the main cards of the cattle the authenticated player has activated, derived from the player's own activation records and isolated to the current player. The optional category filter narrows the result to one card category. Cards are batch-assembled from the activated cattle to avoid N+1. Requires a valid player token."`
	Category string `json:"category" dc:"Optional card category filter: person=人物, event=事件, research=科研, college=院系, spirit=精神; empty returns all owned cards" eg:"spirit"`
}

// CollectionRes is the response for the current player's personal card collection.
type CollectionRes struct {
	List []*CollectionCardItem `json:"list" dc:"Owned cards ordered by activation recency (most recent first)" eg:"[]"`
}

// CollectionCardItem defines one collected card with its owning cattle reference.
type CollectionCardItem struct {
	NiuId     int64  `json:"niuId" dc:"Owning cattle ID" eg:"1"`
	NiuCode   string `json:"niuCode" dc:"Owning cattle serial code" eg:"NIU-001"`
	Category  string `json:"category" dc:"Card category: person=人物, event=事件, research=科研, college=院系, spirit=精神" eg:"spirit"`
	Title     string `json:"title" dc:"Card title" eg:"川农大精神"`
	Content   string `json:"content" dc:"Card content text" eg:"爱国敬业、艰苦奋斗、团结拼搏、求实创新"`
	ImagePath string `json:"imagePath" dc:"Card image storage path; empty when none" eg:"sicau-niu/card/1.jpg"`
}
