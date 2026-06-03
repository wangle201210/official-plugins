// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"time"

	"github.com/gogf/gf/v2/frame/g"
)

// Settlement is the golang structure of table plugin_sicau_niu_settlement for DAO operations like Where/Data.
type Settlement struct {
	g.Meta     `orm:"table:plugin_sicau_niu_settlement, do:true"`
	Id         any        //
	Title      any        // Archive title given by the operator
	Snapshot   any        // Frozen dashboard metrics serialized as JSON text
	OperatorId any        // Host operator user ID who created the archive
	ArchivedAt *time.Time // Archive time, set explicitly by the settlement service
	CreatedAt  *time.Time // Creation time
	UpdatedAt  *time.Time // Update time
	DeletedAt  *time.Time // Soft-delete time, NULL means active
}
