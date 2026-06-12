// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"time"

	"github.com/gogf/gf/v2/frame/g"
)

// Steal is the golang structure of table plugin_sicau_niu_steal for DAO operations like Where/Data.
type Steal struct {
	g.Meta       `orm:"table:plugin_sicau_niu_steal, do:true"`
	Id           any        //
	ActorUserId  any        //
	TargetUserId any        //
	Amount       any        //
	StealDate    any        // Steal date YYYY-MM-DD for the daily count limit
	CreatedAt    *time.Time //
	UpdatedAt    *time.Time //
	DeletedAt    *time.Time //
	RequestId    any        // Client idempotency key deduplicating network retries; empty when not provided
}
