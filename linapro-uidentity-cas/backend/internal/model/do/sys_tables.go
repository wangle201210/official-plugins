// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"time"

	"github.com/gogf/gf/v2/frame/g"
)

// SysTables is the golang structure of table plugin_linapro_uidentity_cas_sys_tables for DAO operations like Where/Data.
type SysTables struct {
	g.Meta              `orm:"table:plugin_linapro_uidentity_cas_sys_tables, do:true"`
	TableId             any        //
	TableName           any        //
	TableComment        any        //
	ClassName           any        //
	TplCategory         any        //
	PackageName         any        //
	ModuleName          any        //
	ModuleFrontName     any        //
	BusinessName        any        //
	FunctionName        any        //
	FunctionAuthor      any        //
	PkColumn            any        //
	PkGoField           any        //
	PkJsonField         any        //
	Options             any        //
	TreeCode            any        //
	TreeParentCode      any        //
	TreeName            any        //
	Tree                any        //
	Crud                any        //
	Remark              any        //
	IsDataScope         any        //
	IsActions           any        //
	IsAuth              any        //
	IsLogicalDelete     any        //
	LogicalDelete       any        //
	LogicalDeleteColumn any        //
	CreateBy            any        //
	UpdateBy            any        //
	CreatedAt           *time.Time //
	UpdatedAt           *time.Time //
	DeletedAt           *time.Time //
}
