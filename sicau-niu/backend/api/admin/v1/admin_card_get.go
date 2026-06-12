// admin_card_get.go defines the request and response DTOs for fetching one card
// detail by ID.

package v1

import "github.com/gogf/gf/v2/frame/g"

// GetCardReq is the request for fetching one card detail.
type GetCardReq struct {
	g.Meta `path:"/plugins/sicau-niu/admin/cards/{id}" method:"get" tags:"Sicau Niu Admin" summary:"获取卡片详情" dc:"Fetch one card detail by ID, including the batch-assembled owning-cattle code and name. Protected by host unified permission check." permission:"sicau-niu:card:list"`
	Id     int64 `json:"id" v:"required|min:1" dc:"Card ID from the path" eg:"1"`
}

// GetCardRes is the response for fetching one card detail.
type GetCardRes struct {
	*CardItem
}
