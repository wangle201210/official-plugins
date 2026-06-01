// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"time"

	"github.com/gogf/gf/v2/frame/g"
)

// Unit is the golang structure of table plugin_linapro_uidentity_cas_unit for DAO operations like Where/Data.
type Unit struct {
	g.Meta    `orm:"table:plugin_linapro_uidentity_cas_unit, do:true"`
	Id        any        //
	TenantId  any        //
	Name      any        //
	Alias     any        //
	Code      any        //
	CreatedBy any        //
	UpdatedBy any        //
	CreatedAt *time.Time //
	UpdatedAt *time.Time //
	DeletedAt *time.Time //
}
