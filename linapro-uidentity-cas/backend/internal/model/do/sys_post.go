// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"time"

	"github.com/gogf/gf/v2/frame/g"
)

// SysPost is the golang structure of table plugin_linapro_uidentity_cas_sys_post for DAO operations like Where/Data.
type SysPost struct {
	g.Meta    `orm:"table:plugin_linapro_uidentity_cas_sys_post, do:true"`
	PostId    any        //
	PostName  any        //
	PostCode  any        //
	Sort      any        //
	Status    any        //
	Remark    any        //
	CreateBy  any        //
	UpdateBy  any        //
	CreatedAt *time.Time //
	UpdatedAt *time.Time //
	DeletedAt *time.Time //
}
