// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"time"

	"github.com/gogf/gf/v2/frame/g"
)

// Applications is the golang structure of table plugin_linapro_uidentity_cas_applications for DAO operations like Where/Data.
type Applications struct {
	g.Meta      `orm:"table:plugin_linapro_uidentity_cas_applications, do:true"`
	Id          any        //
	Name        any        //
	Alias       any        //
	ClientId    any        //
	SecretKey   any        //
	AccessModel any        //
	Status      any        //
	CallbackUrl any        //
	Whitelist   any        //
	CreatedAt   *time.Time //
	UpdatedAt   *time.Time //
	DeletedAt   *time.Time //
	CreateBy    any        //
	UpdateBy    any        //
}
