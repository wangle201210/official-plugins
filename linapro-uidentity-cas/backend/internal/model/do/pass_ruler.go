// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"time"

	"github.com/gogf/gf/v2/frame/g"
)

// PassRuler is the golang structure of table plugin_linapro_uidentity_cas_pass_ruler for DAO operations like Where/Data.
type PassRuler struct {
	g.Meta         `orm:"table:plugin_linapro_uidentity_cas_pass_ruler, do:true"`
	Id             any        //
	Name           any        //
	Capital        any        //
	Lower          any        //
	Number         any        //
	Symbol         any        //
	Length         any        //
	Interval       any        //
	IntervalStatus any        //
	Status         any        //
	CreatedAt      *time.Time //
	UpdatedAt      *time.Time //
	DeletedAt      *time.Time //
	CreateBy       any        //
	UpdateBy       any        //
}
