// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

import (
	"time"
)

// PassRule is the golang structure for table pass_rule.
type PassRule struct {
	Id             int64      `json:"id"             orm:"id"              description:""`
	TenantId       int        `json:"tenantId"       orm:"tenant_id"       description:""`
	Name           string     `json:"name"           orm:"name"            description:""`
	Capital        int        `json:"capital"        orm:"capital"         description:""`
	Lower          int        `json:"lower"          orm:"lower"           description:""`
	Number         int        `json:"number"         orm:"number"          description:""`
	Symbol         int        `json:"symbol"         orm:"symbol"          description:""`
	Length         int        `json:"length"         orm:"length"          description:""`
	IntervalDays   int        `json:"intervalDays"   orm:"interval_days"   description:""`
	IntervalStatus int        `json:"intervalStatus" orm:"interval_status" description:""`
	Status         int        `json:"status"         orm:"status"          description:"Rule status: 0=disabled, 1=enabled"`
	CreatedBy      int64      `json:"createdBy"      orm:"created_by"      description:""`
	UpdatedBy      int64      `json:"updatedBy"      orm:"updated_by"      description:""`
	CreatedAt      *time.Time `json:"createdAt"      orm:"created_at"      description:""`
	UpdatedAt      *time.Time `json:"updatedAt"      orm:"updated_at"      description:""`
	DeletedAt      *time.Time `json:"deletedAt"      orm:"deleted_at"      description:""`
}
