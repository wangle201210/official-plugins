// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"time"

	"github.com/gogf/gf/v2/frame/g"
)

// Activation is the golang structure of table plugin_sicau_niu_activation for DAO operations like Where/Data.
type Activation struct {
	g.Meta       `orm:"table:plugin_sicau_niu_activation, do:true"`
	Id           any        //
	UserId       any        // Activating player ID
	NiuId        any        // Activated cattle ID
	ActivityDate any        // Activation date YYYY-MM-DD for the per-day limit
	ActivatedAt  *time.Time // Activation time
	IsFirst      any        // Whether this is the cattle first-activation: 1 first, 0 later
	OrderNo      any        // Arrival order for this cattle, starting at 1
	PhotoPath    any        // Optional activation photo storage path (evidence only, no recognition)
	CreatedAt    *time.Time // Creation time
	UpdatedAt    *time.Time // Update time
	DeletedAt    *time.Time // Soft-delete time, NULL means active
}
