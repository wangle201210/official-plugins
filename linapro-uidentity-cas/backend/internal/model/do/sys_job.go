// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"time"

	"github.com/gogf/gf/v2/frame/g"
)

// SysJob is the golang structure of table plugin_linapro_uidentity_cas_sys_job for DAO operations like Where/Data.
type SysJob struct {
	g.Meta         `orm:"table:plugin_linapro_uidentity_cas_sys_job, do:true"`
	JobId          any        //
	TenantId       any        //
	JobName        any        //
	JobGroup       any        //
	JobType        any        // Job type: 1=http, 2=exec or plugin-defined executor
	CronExpression any        //
	InvokeTarget   any        //
	Args           any        //
	MisfirePolicy  any        // Misfire policy copied from legacy sys_job semantics
	Concurrent     any        // Concurrent execution flag: 0=disallow, 1=allow
	Status         any        // Job status: 1=disabled, 2=enabled
	EntryId        any        // Runtime scheduler entry ID, 0 means not scheduled
	CreatedBy      any        //
	UpdatedBy      any        //
	CreatedAt      *time.Time //
	UpdatedAt      *time.Time //
	DeletedAt      *time.Time //
}
