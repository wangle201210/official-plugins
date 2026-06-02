// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

import (
	"time"
)

// AccountActiveLog is the golang structure for table account_active_log.
type AccountActiveLog struct {
	Id        int64      `json:"id"        orm:"id"         description:""`
	TenantId  int        `json:"tenantId"  orm:"tenant_id"  description:""`
	Number    string     `json:"number"    orm:"number"     description:""`
	Phone     string     `json:"phone"     orm:"phone"      description:""`
	Wechat    string     `json:"wechat"    orm:"wechat"     description:""`
	Type      int        `json:"type"      orm:"type"       description:"Legacy activation log type: 0=activation or Wechat rebind callback, 1=union ID bind"`
	CreatedBy int64      `json:"createdBy" orm:"created_by" description:""`
	UpdatedBy int64      `json:"updatedBy" orm:"updated_by" description:""`
	CreatedAt *time.Time `json:"createdAt" orm:"created_at" description:""`
	UpdatedAt *time.Time `json:"updatedAt" orm:"updated_at" description:""`
	DeletedAt *time.Time `json:"deletedAt" orm:"deleted_at" description:""`
}
