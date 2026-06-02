// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"time"

	"github.com/gogf/gf/v2/frame/g"
)

// SysApi is the golang structure of table plugin_linapro_uidentity_cas_sys_api for DAO operations like Where/Data.
type SysApi struct {
	g.Meta    `orm:"table:plugin_linapro_uidentity_cas_sys_api, do:true"`
	Id        any        //
	Handle    any        //
	Title     any        //
	Path      any        //
	Type      any        //
	Action    any        //
	CreateBy  any        //
	UpdateBy  any        //
	CreatedAt *time.Time //
	UpdatedAt *time.Time //
	DeletedAt *time.Time //
}
