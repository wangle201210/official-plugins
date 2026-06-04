// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"time"

	"github.com/gogf/gf/v2/frame/g"
)

// College is the golang structure of table plugin_sicau_niu_college for DAO operations like Where/Data.
type College struct {
	g.Meta    `orm:"table:plugin_sicau_niu_college, do:true"`
	Id        any        // Primary key ID
	Name      any        // College name
	Sort      any        // Display sort order, smaller first
	CreatedAt *time.Time // Creation time
	UpdatedAt *time.Time // Update time
	DeletedAt *time.Time // Soft-delete time, NULL means active
}
