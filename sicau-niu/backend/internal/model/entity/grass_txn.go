// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

import (
	"time"
)

// GrassTxn is the golang structure for table grass_txn.
type GrassTxn struct {
	Id        int64      `json:"id"        orm:"id"         description:""`
	UserId    int64      `json:"userId"    orm:"user_id"    description:""`
	Delta     int64      `json:"delta"     orm:"delta"      description:"Signed grass change: positive credit, negative debit"`
	TxnType   string     `json:"txnType"   orm:"txn_type"   description:"Type: checkin, feed, steal_gain, stolen_loss, gift_out, gift_in"`
	RefId     int64      `json:"refId"     orm:"ref_id"     description:"Related business record ID (feeding/steal/gift/checkin)"`
	CreatedAt *time.Time `json:"createdAt" orm:"created_at" description:""`
	UpdatedAt *time.Time `json:"updatedAt" orm:"updated_at" description:""`
	DeletedAt *time.Time `json:"deletedAt" orm:"deleted_at" description:""`
}
