// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

import (
	"time"
)

// AccountAppBlacklist is the golang structure for table account_app_blacklist.
type AccountAppBlacklist struct {
	Id        int64      `json:"id"        orm:"id"         description:""`
	Name      string     `json:"name"      orm:"name"       description:""`
	AppId     int64      `json:"appId"     orm:"app_id"     description:""`
	AccountId int64      `json:"accountId" orm:"account_id" description:""`
	EffectAt  *time.Time `json:"effectAt"  orm:"effect_at"  description:""`
	ExpireAt  *time.Time `json:"expireAt"  orm:"expire_at"  description:""`
	CreatedAt *time.Time `json:"createdAt" orm:"created_at" description:""`
	UpdatedAt *time.Time `json:"updatedAt" orm:"updated_at" description:""`
	DeletedAt *time.Time `json:"deletedAt" orm:"deleted_at" description:""`
	CreateBy  int64      `json:"createBy"  orm:"create_by"  description:""`
	UpdateBy  int64      `json:"updateBy"  orm:"update_by"  description:""`
}
