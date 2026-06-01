// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

import (
	"time"
)

// JobLog is the golang structure for table job_log.
type JobLog struct {
	Id        int64      `json:"id"        orm:"id"         description:""`
	TenantId  int        `json:"tenantId"  orm:"tenant_id"  description:""`
	JobId     int64      `json:"jobId"     orm:"job_id"     description:""`
	JobName   string     `json:"jobName"   orm:"job_name"   description:""`
	StartAt   *time.Time `json:"startAt"   orm:"start_at"   description:""`
	EndAt     *time.Time `json:"endAt"     orm:"end_at"     description:""`
	CreateNum int64      `json:"createNum" orm:"create_num" description:""`
	UpdateNum int64      `json:"updateNum" orm:"update_num" description:""`
	DeleteNum int64      `json:"deleteNum" orm:"delete_num" description:""`
	ErrNum    int64      `json:"errNum"    orm:"err_num"    description:""`
	CreatedBy int64      `json:"createdBy" orm:"created_by" description:""`
	UpdatedBy int64      `json:"updatedBy" orm:"updated_by" description:""`
	CreatedAt *time.Time `json:"createdAt" orm:"created_at" description:""`
	UpdatedAt *time.Time `json:"updatedAt" orm:"updated_at" description:""`
	DeletedAt *time.Time `json:"deletedAt" orm:"deleted_at" description:""`
}
