// player_grass.go defines the request and response DTOs for the player's grass
// account view: the current ledger balance plus a bounded page of recent
// transactions (type, signed delta, time). The view is isolated to the
// authenticated player's own account.

package v1

import "github.com/gogf/gf/v2/frame/g"

// GrassAccountReq is the request for the current player's grass account view.
type GrassAccountReq struct {
	g.Meta `path:"/plugins/sicau-niu/player/grass" method:"get" tags:"寻牛小程序" summary:"获取草料余额和近期流水" dc:"Return the authenticated player's current grass balance and a bounded list of recent ledger transactions (most recent first). The account is isolated to the current player. The recent-transaction list size is capped server-side. Requires a valid player token."`
}

// GrassAccountRes is the response for the current player's grass account view.
type GrassAccountRes struct {
	Balance         int64           `json:"balance" dc:"Current grass balance" eg:"133"`
	Level           int             `json:"level" dc:"Player level derived from cumulative effective feeding" eg:"3"`
	Exp             int64           `json:"exp" dc:"Cumulative effective feeding experience" eg:"245"`
	ExpIntoLevel    int64           `json:"expIntoLevel" dc:"Experience already earned inside the current level" eg:"45"`
	ExpForNextLevel int64           `json:"expForNextLevel" dc:"Experience the current level needs in total to advance" eg:"200"`
	CheckedToday    bool            `json:"checkedToday" dc:"Whether the player checked in during the current Beijing natural day" eg:"true"`
	Recent          []*GrassTxnItem `json:"recent" dc:"Recent ledger transactions ordered by time descending (bounded)" eg:"[]"`
}

// GrassTxnItem defines one ledger transaction projected for the player account.
type GrassTxnItem struct {
	TxnType   string `json:"txnType" dc:"Transaction type: checkin=签到, feed=喂草, steal_gain=偷草所得, stolen_loss=被偷损失, gift_out=送出, gift_in=收到赠草" eg:"checkin"`
	Delta     int64  `json:"delta" dc:"Signed grass change: positive credit, negative debit" eg:"33"`
	CreatedAt *int64 `json:"createdAt" dc:"Transaction time as Unix timestamp in milliseconds" eg:"1776333600000"`
}
