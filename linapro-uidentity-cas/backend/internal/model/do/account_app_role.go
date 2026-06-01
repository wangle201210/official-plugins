// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"time"

	"github.com/gogf/gf/v2/frame/g"
)

// AccountAppRole is the golang structure of table plugin_linapro_uidentity_cas_account_app_role for DAO operations like Where/Data.
type AccountAppRole struct {
	g.Meta             `orm:"table:plugin_linapro_uidentity_cas_account_app_role, do:true"`
	Id                 any        //
	TenantId           any        //
	GiveAccountId      any        //
	EmpoweredAccountId any        //
	AppId              any        //
	ExpireAt           *time.Time //
	CreatedBy          any        //
	UpdatedBy          any        //
	CreatedAt          *time.Time //
	UpdatedAt          *time.Time //
	DeletedAt          *time.Time //
}
