// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"time"

	"github.com/gogf/gf/v2/frame/g"
)

// GrassAccount is the golang structure of table plugin_sicau_niu_grass_account for DAO operations like Where/Data.
type GrassAccount struct {
	g.Meta    `orm:"table:plugin_sicau_niu_grass_account, do:true"`
	Id        any        //
	UserId    any        // Owning player ID
	Balance   any        // Current grass balance
	CreatedAt *time.Time //
	UpdatedAt *time.Time //
	DeletedAt *time.Time // Soft-delete time, NULL means active
}
