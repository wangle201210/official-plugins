// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"time"

	"github.com/gogf/gf/v2/frame/g"
)

// SysMenu is the golang structure of table plugin_linapro_uidentity_cas_sys_menu for DAO operations like Where/Data.
type SysMenu struct {
	g.Meta     `orm:"table:plugin_linapro_uidentity_cas_sys_menu, do:true"`
	MenuId     any        //
	MenuName   any        //
	Title      any        //
	Icon       any        //
	Path       any        //
	Paths      any        //
	MenuType   any        //
	Action     any        //
	Permission any        //
	ParentId   any        //
	NoCache    any        //
	Breadcrumb any        //
	Component  any        //
	Sort       any        //
	Visible    any        //
	IsFrame    any        //
	CreateBy   any        //
	UpdateBy   any        //
	CreatedAt  *time.Time //
	UpdatedAt  *time.Time //
	DeletedAt  *time.Time //
}
