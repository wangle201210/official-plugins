// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"time"

	"github.com/gogf/gf/v2/frame/g"
)

// AccountActiveLog is the golang structure of table plugin_linapro_uidentity_cas_account_active_log for DAO operations like Where/Data.
type AccountActiveLog struct {
	g.Meta    `orm:"table:plugin_linapro_uidentity_cas_account_active_log, do:true"`
	Id        any        //
	TenantId  any        //
	Number    any        //
	Phone     any        //
	Wechat    any        //
	Type      any        // Legacy activation log type: 0=activation or Wechat rebind callback, 1=union ID bind
	CreatedBy any        //
	UpdatedBy any        //
	CreatedAt *time.Time //
	UpdatedAt *time.Time //
	DeletedAt *time.Time //
}
