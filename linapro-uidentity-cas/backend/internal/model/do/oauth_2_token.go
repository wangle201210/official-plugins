// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"time"

	"github.com/gogf/gf/v2/frame/g"
)

// Oauth2Token is the golang structure of table plugin_linapro_uidentity_cas_oauth2_token for DAO operations like Where/Data.
type Oauth2Token struct {
	g.Meta    `orm:"table:plugin_linapro_uidentity_cas_oauth2_token, do:true"`
	Id        any        //
	ExpiredAt any        //
	Code      any        //
	Access    any        //
	Refresh   any        //
	Data      any        //
	CreatedAt *time.Time //
	UpdatedAt *time.Time //
	DeletedAt *time.Time //
	CreateBy  any        //
	UpdateBy  any        //
}
