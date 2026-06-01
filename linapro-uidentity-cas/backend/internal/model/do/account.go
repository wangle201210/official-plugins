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
	g.Meta            `orm:"table:plugin_linapro_uidentity_cas_account, do:true"`
	Id                any        //
	TenantId          any        // Owning tenant ID, 0 means platform
	Number            any        // Stable account number
	Name              any        // Account display name
	Phone             any        // Mobile phone number
	PasswordHash      any        // Password hash managed by the plugin
	EffectAt          *time.Time //
	ExpireAt          *time.Time //
	PasswordUpdatedAt *time.Time //
	PassLevel         any        // Password strength level: 0=invalid, higher is stronger
	ContainerId       any        // Container ID
	UnitId            any        // Primary unit ID
	Status            any        // Account status: 0=not active, 1=normal, 2=locked
	CreatedBy         any        //
	UpdatedBy         any        //
	CreatedAt         *time.Time //
	UpdatedAt         *time.Time //
	DeletedAt         *time.Time //
}
