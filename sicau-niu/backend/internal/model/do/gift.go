// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"time"

	"github.com/gogf/gf/v2/frame/g"
)

// Gift is the golang structure of table plugin_sicau_niu_gift for DAO operations like Where/Data.
type Gift struct {
	g.Meta        `orm:"table:plugin_sicau_niu_gift, do:true"`
	Id            any        //
	FromUserId    any        //
	ToUserId      any        //
	Amount        any        //
	GiftDate      any        // Gift date YYYY-MM-DD for the daily count limit
	CreatedAt     *time.Time //
	UpdatedAt     *time.Time //
	DeletedAt     *time.Time //
	RequestId     any        // Required player-scoped idempotency key for stable replay
	ResultBalance any        // Giver balance returned by the first successful gift request
}
