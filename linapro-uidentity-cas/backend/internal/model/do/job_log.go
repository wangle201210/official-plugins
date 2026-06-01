// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"time"

	"github.com/gogf/gf/v2/frame/g"
)

// JobLog is the golang structure of table plugin_linapro_uidentity_cas_job_log for DAO operations like Where/Data.
type JobLog struct {
	g.Meta    `orm:"table:plugin_linapro_uidentity_cas_job_log, do:true"`
	Id        any        //
	TenantId  any        //
	JobId     any        //
	JobName   any        //
	StartAt   *time.Time //
	EndAt     *time.Time //
	CreateNum any        //
	UpdateNum any        //
	DeleteNum any        //
	ErrNum    any        //
	CreatedBy any        //
	UpdatedBy any        //
	CreatedAt *time.Time //
	UpdatedAt *time.Time //
	DeletedAt *time.Time //
}
