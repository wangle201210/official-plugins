// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"time"

	"github.com/gogf/gf/v2/frame/g"
)

// SysDictData is the golang structure of table plugin_linapro_uidentity_cas_sys_dict_data for DAO operations like Where/Data.
type SysDictData struct {
	g.Meta    `orm:"table:plugin_linapro_uidentity_cas_sys_dict_data, do:true"`
	DictCode  any        //
	DictSort  any        //
	DictLabel any        //
	DictValue any        //
	DictType  any        //
	CssClass  any        //
	ListClass any        //
	IsDefault any        //
	Status    any        //
	Default   any        //
	Remark    any        //
	CreateBy  any        //
	UpdateBy  any        //
	CreatedAt *time.Time //
	UpdatedAt *time.Time //
	DeletedAt *time.Time //
}
