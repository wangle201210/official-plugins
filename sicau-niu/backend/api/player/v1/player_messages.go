// player_messages.go defines the request and response DTOs for the player's
// in-app inbox: a paged list of the player's own messages (GET) and marking one
// message read (PUT). Messages are produced by social events (stolen, gift
// received) and are isolated to the authenticated player.

package v1

import "github.com/gogf/gf/v2/frame/g"

// MessagesReq is the request for the current player's paged inbox.
type MessagesReq struct {
	g.Meta   `path:"/plugins/sicau-niu/player/messages" method:"get" tags:"Sicau Niu Player" summary:"List in-app messages" dc:"Return the authenticated player's in-app messages, newest first, with pagination. The inbox is isolated to the current player. Requires a valid player token."`
	PageNum  int `json:"pageNum" dc:"Page number; defaults to 1 when omitted or non-positive" eg:"1"`
	PageSize int `json:"pageSize" dc:"Page size; defaults to 10 and is capped at 100" eg:"10"`
}

// MessagesRes is the response for the current player's paged inbox.
type MessagesRes struct {
	List  []*MessageItem `json:"list" dc:"In-app messages ordered by time descending" eg:"[]"`
	Total int            `json:"total" dc:"Total message count for the player" eg:"3"`
}

// MessageItem defines one in-app message projected for the player.
type MessageItem struct {
	Id        int64  `json:"id" dc:"Message ID" eg:"1"`
	MsgType   string `json:"msgType" dc:"Message type: stolen=被偷通知, gift_received=收到赠草" eg:"stolen"`
	Content   string `json:"content" dc:"Message content text" eg:"你的草被偷走了 12 份"`
	IsRead    bool   `json:"isRead" dc:"Whether the message is read" eg:"false"`
	CreatedAt *int64 `json:"createdAt" dc:"Message time as Unix timestamp in milliseconds" eg:"1776333600000"`
}

// MarkMessageReadReq is the request for marking one of the player's messages read.
type MarkMessageReadReq struct {
	g.Meta `path:"/plugins/sicau-niu/player/messages/{id}/read" method:"put" tags:"Sicau Niu Player" summary:"Mark a message read" dc:"Mark one of the authenticated player's own in-app messages as read. Marking a message that does not belong to the current player is rejected. Requires a valid player token."`
	Id     int64 `json:"id" v:"required" dc:"Target message ID owned by the current player" eg:"1"`
}

// MarkMessageReadRes is the empty response for a successful mark-read.
type MarkMessageReadRes struct{}
