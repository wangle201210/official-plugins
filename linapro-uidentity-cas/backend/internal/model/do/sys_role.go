// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"time"

	"github.com/gogf/gf/v2/frame/g"
)

// SysRole is the golang structure of table plugin_linapro_uidentity_cas_sys_role for DAO operations like Where/Data.
type SysRole struct {
	g.Meta    `orm:"table:plugin_linapro_uidentity_cas_sys_role, do:true"`
	RoleId    any        //
	RoleName  any        //
	Status    any        //
	RoleKey   any        //
	RoleSort  any        //
	Flag      any        //
	Remark    any        //
	Admin     any        //
	DataScope any        //
	CreateBy  any        //
	UpdateBy  any        //
	CreatedAt *time.Time //
	UpdatedAt *time.Time //
	DeletedAt *time.Time //
}
