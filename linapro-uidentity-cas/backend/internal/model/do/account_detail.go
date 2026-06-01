// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"time"

	"github.com/gogf/gf/v2/frame/g"
)

// AccountDetail is the golang structure of table plugin_linapro_uidentity_cas_account_detail for DAO operations like Where/Data.
type AccountDetail struct {
	g.Meta       `orm:"table:plugin_linapro_uidentity_cas_account_detail, do:true"`
	AccountId    any        //
	TenantId     any        //
	Birthday     any        // Date-only birthday in YYYY-MM-DD format
	Email        any        //
	Gender       any        //
	Qq           any        //
	Wechat       any        //
	Idcard       any        //
	Avatar       any        //
	Source       any        //
	Grade        any        //
	College      any        //
	CollegeCode  any        //
	Campus       any        //
	SchoolSystem any        //
	GraduatedAt  any        //
	Major        any        //
	ClassName    any        //
	Face         any        //
	CreatedBy    any        //
	UpdatedBy    any        //
	CreatedAt    *time.Time //
	UpdatedAt    *time.Time //
}
