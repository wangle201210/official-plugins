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
	g.Meta     `orm:"table:plugin_sicau_niu_gift, do:true"`
	Id         any        //
	FromUserId any        //
	ToUserId   any        //
	Amount     any        //
	GiftDate   any        // Gift date YYYY-MM-DD for the daily count limit
	CreatedAt  *time.Time //
	UpdatedAt  *time.Time //
	DeletedAt  *time.Time //
	RequestId  any        // Client idempotency key deduplicating network retries; empty when not provided
}
