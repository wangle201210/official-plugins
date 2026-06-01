// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"time"

	"github.com/gogf/gf/v2/frame/g"
)

// Container is the golang structure of table plugin_linapro_uidentity_cas_container for DAO operations like Where/Data.
type Container struct {
	g.Meta       `orm:"table:plugin_linapro_uidentity_cas_container, do:true"`
	Id           any        //
	TenantId     any        //
	Name         any        //
	Alias        any        //
	AccountCount any        //
	AdminCount   any        //
	CreatedBy    any        //
	UpdatedBy    any        //
	CreatedAt    *time.Time //
	UpdatedAt    *time.Time //
	DeletedAt    *time.Time //
}
