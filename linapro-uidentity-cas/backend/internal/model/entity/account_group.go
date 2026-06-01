// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

import (
	"time"
)

// AccountGroup is the golang structure for table account_group.
type AccountGroup struct {
	Id        int64      `json:"id"        orm:"id"         description:""`
	TenantId  int        `json:"tenantId"  orm:"tenant_id"  description:""`
	AccountId int64      `json:"accountId" orm:"account_id" description:""`
	GroupId   int64      `json:"groupId"   orm:"group_id"   description:""`
	CreatedBy int64      `json:"createdBy" orm:"created_by" description:""`
	CreatedAt *time.Time `json:"createdAt" orm:"created_at" description:""`
	UpdatedAt *time.Time `json:"updatedAt" orm:"updated_at" description:""`
}
