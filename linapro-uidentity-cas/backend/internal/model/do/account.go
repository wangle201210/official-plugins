// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"time"

	"github.com/gogf/gf/v2/frame/g"
)

// Account is the golang structure of table plugin_linapro_uidentity_cas_account for DAO operations like Where/Data.
type Account struct {
	g.Meta      `orm:"table:plugin_linapro_uidentity_cas_account, do:true"`
	Id          any        //
	Number      any        //
	Name        any        //
	Phone       any        //
	EffectAt    *time.Time //
	ExpireAt    *time.Time //
	GroupId     any        //
	PassLevel   any        //
	ContainerId any        //
	UnitId      any        //
	Status      any        //
	CreatedAt   *time.Time //
	UpdatedAt   *time.Time //
	DeletedAt   *time.Time //
	CreateBy    any        //
	UpdateBy    any        //
}
