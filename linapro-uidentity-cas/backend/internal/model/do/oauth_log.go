// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"time"

	"github.com/gogf/gf/v2/frame/g"
)

// OauthLog is the golang structure of table plugin_linapro_uidentity_cas_oauth_log for DAO operations like Where/Data.
type OauthLog struct {
	g.Meta      `orm:"table:plugin_linapro_uidentity_cas_oauth_log, do:true"`
	Id          any        //
	TenantId    any        //
	UserId      any        //
	AppId       any        //
	RedirectUri any        //
	Scope       any        //
	CreatedBy   any        //
	UpdatedBy   any        //
	CreatedAt   *time.Time //
	UpdatedAt   *time.Time //
	DeletedAt   *time.Time //
}
