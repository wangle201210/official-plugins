// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"time"

	"github.com/gogf/gf/v2/frame/g"
)

// Units is the golang structure of table plugin_linapro_uidentity_cas_units for DAO operations like Where/Data.
type Units struct {
	g.Meta    `orm:"table:plugin_linapro_uidentity_cas_units, do:true"`
	Id        any        //
	Name      any        //
	Alias     any        //
	Code      any        //
	CreatedAt *time.Time //
	UpdatedAt *time.Time //
	DeletedAt *time.Time //
	CreateBy  any        //
	UpdateBy  any        //
}
