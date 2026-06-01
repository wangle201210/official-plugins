// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"time"

	"github.com/gogf/gf/v2/frame/g"
)

// GroupAppBlacklist is the golang structure of table plugin_linapro_uidentity_cas_group_app_blacklist for DAO operations like Where/Data.
type GroupAppBlacklist struct {
	g.Meta    `orm:"table:plugin_linapro_uidentity_cas_group_app_blacklist, do:true"`
	Id        any        //
	TenantId  any        //
	Name      any        //
	AppId     any        //
	GroupId   any        //
	EffectAt  *time.Time //
	ExpireAt  *time.Time //
	CreatedBy any        //
	UpdatedBy any        //
	CreatedAt *time.Time //
	UpdatedAt *time.Time //
	DeletedAt *time.Time //
}
