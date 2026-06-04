// admin_card_delete.go defines the request and response DTOs for deleting one
// card.

package v1

import "github.com/gogf/gf/v2/frame/g"

// DeleteCardReq is the request for deleting one card.
type DeleteCardReq struct {
	g.Meta `path:"/plugins/sicau-niu/admin/cards/{id}" method:"delete" tags:"Sicau Niu Admin" summary:"Delete card" dc:"Soft-delete a card, freeing its owning cattle to be bound again. Recoverable. Protected by host unified permission check." permission:"sicau-niu:card:delete"`
	Id     int64 `json:"id" v:"required|min:1" dc:"Card ID from the path" eg:"1"`
}

// DeleteCardRes is the response for deleting one card. It is intentionally empty;
// success is conveyed by the absence of an error.
type DeleteCardRes struct{}
