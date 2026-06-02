// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"time"

	"github.com/gogf/gf/v2/frame/g"
)

// SysConfig is the golang structure of table plugin_linapro_uidentity_cas_sys_config for DAO operations like Where/Data.
type SysConfig struct {
	g.Meta      `orm:"table:plugin_linapro_uidentity_cas_sys_config, do:true"`
	Id          any        //
	ConfigName  any        //
	ConfigKey   any        //
	ConfigValue any        //
	ConfigType  any        //
	IsFrontend  any        //
	Remark      any        //
	CreateBy    any        //
	UpdateBy    any        //
	CreatedAt   *time.Time //
	UpdatedAt   *time.Time //
	DeletedAt   *time.Time //
}
