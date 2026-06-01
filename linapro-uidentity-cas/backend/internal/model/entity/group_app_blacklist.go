// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

import (
	"time"
)

// GroupAppBlacklist is the golang structure for table group_app_blacklist.
type GroupAppBlacklist struct {
	Id        int64      `json:"id"        orm:"id"         description:""`
	TenantId  int        `json:"tenantId"  orm:"tenant_id"  description:""`
	Name      string     `json:"name"      orm:"name"       description:""`
	AppId     int64      `json:"appId"     orm:"app_id"     description:""`
	GroupId   int64      `json:"groupId"   orm:"group_id"   description:""`
	EffectAt  *time.Time `json:"effectAt"  orm:"effect_at"  description:""`
	ExpireAt  *time.Time `json:"expireAt"  orm:"expire_at"  description:""`
	CreatedBy int64      `json:"createdBy" orm:"created_by" description:""`
	UpdatedBy int64      `json:"updatedBy" orm:"updated_by" description:""`
	CreatedAt *time.Time `json:"createdAt" orm:"created_at" description:""`
	UpdatedAt *time.Time `json:"updatedAt" orm:"updated_at" description:""`
	DeletedAt *time.Time `json:"deletedAt" orm:"deleted_at" description:""`
}
