// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"time"

	"github.com/gogf/gf/v2/frame/g"
)

// SysLoginLog is the golang structure of table plugin_linapro_uidentity_cas_sys_login_log for DAO operations like Where/Data.
type SysLoginLog struct {
	g.Meta        `orm:"table:plugin_linapro_uidentity_cas_sys_login_log, do:true"`
	Id            any        //
	Username      any        //
	Status        any        //
	Ipaddr        any        //
	LoginLocation any        //
	Browser       any        //
	Os            any        //
	Platform      any        //
	LoginTime     *time.Time //
	Remark        any        //
	Msg           any        //
	CreatedAt     *time.Time //
	UpdatedAt     *time.Time //
	CreateBy      any        //
	UpdateBy      any        //
}
