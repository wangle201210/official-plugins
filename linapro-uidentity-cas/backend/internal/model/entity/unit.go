// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

import (
	"time"
)

// Unit is the golang structure for table unit.
type Unit struct {
	Id        int64      `json:"id"        orm:"id"         description:""`
	TenantId  int        `json:"tenantId"  orm:"tenant_id"  description:""`
	Name      string     `json:"name"      orm:"name"       description:""`
	Alias     string     `json:"alias"     orm:"alias"      description:""`
	Code      string     `json:"code"      orm:"code"       description:""`
	CreatedBy int64      `json:"createdBy" orm:"created_by" description:""`
	UpdatedBy int64      `json:"updatedBy" orm:"updated_by" description:""`
	CreatedAt *time.Time `json:"createdAt" orm:"created_at" description:""`
	UpdatedAt *time.Time `json:"updatedAt" orm:"updated_at" description:""`
	DeletedAt *time.Time `json:"deletedAt" orm:"deleted_at" description:""`
}
