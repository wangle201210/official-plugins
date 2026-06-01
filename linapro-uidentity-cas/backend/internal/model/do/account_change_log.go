// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"time"

	"github.com/gogf/gf/v2/frame/g"
)

// AccountChangeLog is the golang structure of table plugin_linapro_uidentity_cas_account_change_log for DAO operations like Where/Data.
type AccountChangeLog struct {
	g.Meta    `orm:"table:plugin_linapro_uidentity_cas_account_change_log, do:true"`
	Id        any        //
	TenantId  any        //
	AccountId any        //
	TableName any        //
	Action    any        //
	DataOld   any        //
	DataNew   any        //
	ErrMsg    any        //
	ErrNumber any        //
	CreatedBy any        //
	UpdatedBy any        //
	CreatedAt *time.Time //
	UpdatedAt *time.Time //
	DeletedAt *time.Time //
}
