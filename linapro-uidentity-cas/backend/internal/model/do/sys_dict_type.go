// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"time"

	"github.com/gogf/gf/v2/frame/g"
)

// SysDictType is the golang structure of table plugin_linapro_uidentity_cas_sys_dict_type for DAO operations like Where/Data.
type SysDictType struct {
	g.Meta    `orm:"table:plugin_linapro_uidentity_cas_sys_dict_type, do:true"`
	DictId    any        //
	DictName  any        //
	DictType  any        //
	Status    any        //
	Remark    any        //
	CreateBy  any        //
	UpdateBy  any        //
	CreatedAt *time.Time //
	UpdatedAt *time.Time //
	DeletedAt *time.Time //
}
