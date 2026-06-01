// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"time"

	"github.com/gogf/gf/v2/frame/g"
)

// Application is the golang structure of table plugin_linapro_uidentity_cas_application for DAO operations like Where/Data.
type Application struct {
	g.Meta      `orm:"table:plugin_linapro_uidentity_cas_application, do:true"`
	Id          any        //
	TenantId    any        //
	Name        any        //
	Alias       any        //
	ClientId    any        //
	SecretKey   any        //
	AccessModel any        // Application access model, for example cas/oauth/ldap
	Status      any        // Application status: 0=disabled, 1=enabled
	CallbackUrl any        //
	Whitelist   any        //
	CreatedBy   any        //
	UpdatedBy   any        //
	CreatedAt   *time.Time //
	UpdatedAt   *time.Time //
	DeletedAt   *time.Time //
}
