// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"time"

	"github.com/gogf/gf/v2/frame/g"
)

// OauthToken is the golang structure of table plugin_linapro_uidentity_cas_oauth_token for DAO operations like Where/Data.
type OauthToken struct {
	g.Meta    `orm:"table:plugin_linapro_uidentity_cas_oauth_token, do:true"`
	Id        any        //
	TenantId  any        //
	ExpiredAt *time.Time //
	Code      any        //
	Access    any        //
	Refresh   any        //
	Data      any        //
	CreatedBy any        //
	UpdatedBy any        //
	CreatedAt *time.Time //
	UpdatedAt *time.Time //
	DeletedAt *time.Time //
}
