// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

import (
	"time"
)

// Gift is the golang structure for table gift.
type Gift struct {
	Id            int64      `json:"id"            orm:"id"             description:""`
	FromUserId    int64      `json:"fromUserId"    orm:"from_user_id"   description:""`
	ToUserId      int64      `json:"toUserId"      orm:"to_user_id"     description:""`
	Amount        int        `json:"amount"        orm:"amount"         description:""`
	GiftDate      string     `json:"giftDate"      orm:"gift_date"      description:"Gift date YYYY-MM-DD for the daily count limit"`
	CreatedAt     *time.Time `json:"createdAt"     orm:"created_at"     description:""`
	UpdatedAt     *time.Time `json:"updatedAt"     orm:"updated_at"     description:""`
	DeletedAt     *time.Time `json:"deletedAt"     orm:"deleted_at"     description:""`
	RequestId     string     `json:"requestId"     orm:"request_id"     description:"Required player-scoped idempotency key for stable replay"`
	ResultBalance int64      `json:"resultBalance" orm:"result_balance" description:"Giver balance returned by the first successful gift request"`
}
