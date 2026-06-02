// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"time"

	"github.com/gogf/gf/v2/frame/g"
)

// SysDept is the golang structure of table plugin_linapro_uidentity_cas_sys_dept for DAO operations like Where/Data.
type SysDept struct {
	g.Meta    `orm:"table:plugin_linapro_uidentity_cas_sys_dept, do:true"`
	DeptId    any        //
	ParentId  any        //
	DeptPath  any        //
	DeptName  any        //
	Sort      any        //
	Leader    any        //
	Phone     any        //
	Email     any        //
	Status    any        //
	CreateBy  any        //
	UpdateBy  any        //
	CreatedAt *time.Time //
	UpdatedAt *time.Time //
	DeletedAt *time.Time //
}
