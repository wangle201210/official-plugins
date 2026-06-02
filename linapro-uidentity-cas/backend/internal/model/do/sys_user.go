// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"time"

	"github.com/gogf/gf/v2/frame/g"
)

// SysUser is the golang structure of table plugin_linapro_uidentity_cas_sys_user for DAO operations like Where/Data.
type SysUser struct {
	g.Meta    `orm:"table:plugin_linapro_uidentity_cas_sys_user, do:true"`
	UserId    any        //
	Username  any        //
	Password  any        //
	NickName  any        //
	Phone     any        //
	RoleId    any        //
	Salt      any        //
	Avatar    any        //
	Sex       any        //
	Email     any        //
	DeptId    any        //
	PostId    any        //
	Remark    any        //
	Status    any        //
	CreateBy  any        //
	UpdateBy  any        //
	CreatedAt *time.Time //
	UpdatedAt *time.Time //
	DeletedAt *time.Time //
}
