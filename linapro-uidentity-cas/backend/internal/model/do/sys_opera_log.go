// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"time"

	"github.com/gogf/gf/v2/frame/g"
)

// SysOperaLog is the golang structure of table plugin_linapro_uidentity_cas_sys_opera_log for DAO operations like Where/Data.
type SysOperaLog struct {
	g.Meta        `orm:"table:plugin_linapro_uidentity_cas_sys_opera_log, do:true"`
	Id            any        //
	Title         any        //
	BusinessType  any        //
	BusinessTypes any        //
	Method        any        //
	RequestMethod any        //
	OperatorType  any        //
	OperName      any        //
	DeptName      any        //
	OperUrl       any        //
	OperIp        any        //
	OperLocation  any        //
	OperParam     any        //
	Status        any        //
	OperTime      *time.Time //
	JsonResult    any        //
	Remark        any        //
	LatencyTime   any        //
	UserAgent     any        //
	CreatedAt     *time.Time //
	UpdatedAt     *time.Time //
	CreateBy      any        //
	UpdateBy      any        //
}
