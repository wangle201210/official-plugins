// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"time"

	"github.com/gogf/gf/v2/frame/g"
)

// Containers is the golang structure of table plugin_linapro_uidentity_cas_containers for DAO operations like Where/Data.
type Containers struct {
	g.Meta       `orm:"table:plugin_linapro_uidentity_cas_containers, do:true"`
	Id           any        //
	Name         any        //
	Alias        any        //
	AccountCount any        //
	AdminCount   any        //
	CreatedAt    *time.Time //
	UpdatedAt    *time.Time //
	DeletedAt    *time.Time //
	CreateBy     any        //
	UpdateBy     any        //
}
