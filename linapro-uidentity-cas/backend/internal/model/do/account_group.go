// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"time"

	"github.com/gogf/gf/v2/frame/g"
)

// AccountGroup is the golang structure of table plugin_linapro_uidentity_cas_account_group for DAO operations like Where/Data.
type AccountGroup struct {
	g.Meta    `orm:"table:plugin_linapro_uidentity_cas_account_group, do:true"`
	Id        any        //
	TenantId  any        //
	AccountId any        //
	GroupId   any        //
	CreatedBy any        //
	CreatedAt *time.Time //
	UpdatedAt *time.Time //
}
