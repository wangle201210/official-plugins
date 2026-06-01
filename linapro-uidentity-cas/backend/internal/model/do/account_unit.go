// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"time"

	"github.com/gogf/gf/v2/frame/g"
)

// AccountUnit is the golang structure of table plugin_linapro_uidentity_cas_account_unit for DAO operations like Where/Data.
type AccountUnit struct {
	g.Meta    `orm:"table:plugin_linapro_uidentity_cas_account_unit, do:true"`
	Id        any        //
	TenantId  any        //
	AccountId any        //
	UnitId    any        //
	CreatedBy any        //
	CreatedAt *time.Time //
	UpdatedAt *time.Time //
}
