// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"time"

	"github.com/gogf/gf/v2/frame/g"
)

// GrassTxn is the golang structure of table plugin_sicau_niu_grass_txn for DAO operations like Where/Data.
type GrassTxn struct {
	g.Meta    `orm:"table:plugin_sicau_niu_grass_txn, do:true"`
	Id        any        //
	UserId    any        //
	Delta     any        // Signed grass change: positive credit, negative debit
	TxnType   any        // Type: checkin, feed, steal_gain, stolen_loss, gift_out, gift_in
	RefId     any        // Related business record ID (feeding/steal/gift/checkin)
	CreatedAt *time.Time //
	UpdatedAt *time.Time //
	DeletedAt *time.Time //
}
