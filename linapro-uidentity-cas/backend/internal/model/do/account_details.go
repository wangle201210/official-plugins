// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"time"

	"github.com/gogf/gf/v2/frame/g"
)

// AccountDetails is the golang structure of table plugin_linapro_uidentity_cas_account_details for DAO operations like Where/Data.
type AccountDetails struct {
	g.Meta    `orm:"table:plugin_linapro_uidentity_cas_account_details, do:true"`
	AccountId any        //
	Birthday  *time.Time //
	Email     any        //
	Gender    any        //
	Qq        any        //
	Wechat    any        //
	Idcard    any        //
	Avatar    any        //
	Source    any        //
	Nj        any        //
	Xymc      any        //
	Xydm      any        //
	Xq        any        //
	Xz        any        //
	Yjbysj    any        //
	Zymc      any        //
	Bjmc      any        //
	Face      any        //
	CreatedAt *time.Time //
	UpdatedAt *time.Time //
	DeletedAt *time.Time //
	CreateBy  any        //
	UpdateBy  any        //
}
