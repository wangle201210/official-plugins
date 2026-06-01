// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

import (
	"time"
)

// AccountUnit is the golang structure for table account_unit.
type AccountUnit struct {
	Id        int64      `json:"id"        orm:"id"         description:""`
	TenantId  int        `json:"tenantId"  orm:"tenant_id"  description:""`
	AccountId int64      `json:"accountId" orm:"account_id" description:""`
	UnitId    int64      `json:"unitId"    orm:"unit_id"    description:""`
	CreatedBy int64      `json:"createdBy" orm:"created_by" description:""`
	CreatedAt *time.Time `json:"createdAt" orm:"created_at" description:""`
	UpdatedAt *time.Time `json:"updatedAt" orm:"updated_at" description:""`
}
