// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

import (
	"time"
)

// SysJob is the golang structure for table sys_job.
type SysJob struct {
	JobId          int64      `json:"jobId"          orm:"job_id"          description:""`
	TenantId       int        `json:"tenantId"       orm:"tenant_id"       description:""`
	JobName        string     `json:"jobName"        orm:"job_name"        description:""`
	JobGroup       string     `json:"jobGroup"       orm:"job_group"       description:""`
	JobType        int        `json:"jobType"        orm:"job_type"        description:"Job type: 1=http, 2=exec or plugin-defined executor"`
	CronExpression string     `json:"cronExpression" orm:"cron_expression" description:""`
	InvokeTarget   string     `json:"invokeTarget"   orm:"invoke_target"   description:""`
	Args           string     `json:"args"           orm:"args"            description:""`
	MisfirePolicy  int        `json:"misfirePolicy"  orm:"misfire_policy"  description:"Misfire policy copied from legacy sys_job semantics"`
	Concurrent     int        `json:"concurrent"     orm:"concurrent"      description:"Concurrent execution flag: 0=disallow, 1=allow"`
	Status         int        `json:"status"         orm:"status"          description:"Job status: 1=disabled, 2=enabled"`
	EntryId        int64      `json:"entryId"        orm:"entry_id"        description:"Runtime scheduler entry ID, 0 means not scheduled"`
	CreatedBy      int64      `json:"createdBy"      orm:"created_by"      description:""`
	UpdatedBy      int64      `json:"updatedBy"      orm:"updated_by"      description:""`
	CreatedAt      *time.Time `json:"createdAt"      orm:"created_at"      description:""`
	UpdatedAt      *time.Time `json:"updatedAt"      orm:"updated_at"      description:""`
	DeletedAt      *time.Time `json:"deletedAt"      orm:"deleted_at"      description:""`
}
