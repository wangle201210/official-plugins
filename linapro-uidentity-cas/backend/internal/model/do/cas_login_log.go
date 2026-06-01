// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"time"

	"github.com/gogf/gf/v2/frame/g"
)

// CasLoginLog is the golang structure of table plugin_linapro_uidentity_cas_cas_login_log for DAO operations like Where/Data.
type CasLoginLog struct {
	g.Meta          `orm:"table:plugin_linapro_uidentity_cas_cas_login_log, do:true"`
	Id              any        //
	TenantId        any        //
	AccountId       any        //
	ChoiceAccountId any        //
	AppId           any        //
	Ipaddr          any        //
	LoginLocation   any        //
	Browser         any        //
	Os              any        //
	Platform        any        //
	LoginTime       *time.Time //
	Remark          any        //
	Msg             any        //
	LoginType       any        //
	CreatedBy       any        //
	UpdatedBy       any        //
	CreatedAt       *time.Time //
	UpdatedAt       *time.Time //
	DeletedAt       *time.Time //
}
