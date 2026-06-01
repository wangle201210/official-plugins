// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"time"

	"github.com/gogf/gf/v2/frame/g"
)

// Group is the golang structure of table plugin_linapro_uidentity_cas_group for DAO operations like Where/Data.
type Group struct {
	g.Meta    `orm:"table:plugin_linapro_uidentity_cas_group, do:true"`
	Id        any        //
	TenantId  any        //
	Name      any        //
	Alias     any        //
	CreatedBy any        //
	UpdatedBy any        //
	CreatedAt *time.Time //
	UpdatedAt *time.Time //
	DeletedAt *time.Time //
}
