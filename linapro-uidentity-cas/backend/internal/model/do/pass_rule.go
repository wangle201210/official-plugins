// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"time"

	"github.com/gogf/gf/v2/frame/g"
)

// PassRule is the golang structure of table plugin_linapro_uidentity_cas_pass_rule for DAO operations like Where/Data.
type PassRule struct {
	g.Meta         `orm:"table:plugin_linapro_uidentity_cas_pass_rule, do:true"`
	Id             any        //
	TenantId       any        //
	Name           any        //
	Capital        any        //
	Lower          any        //
	Number         any        //
	Symbol         any        //
	Length         any        //
	IntervalDays   any        //
	IntervalStatus any        //
	Status         any        // Rule status: 0=disabled, 1=enabled
	CreatedBy      any        //
	UpdatedBy      any        //
	CreatedAt      *time.Time //
	UpdatedAt      *time.Time //
	DeletedAt      *time.Time //
}
