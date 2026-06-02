// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

import (
	"time"
)

// Account is the golang structure for table account.
type Account struct {
	Id          int64      `json:"id"          orm:"id"           description:""`
	Number      string     `json:"number"      orm:"number"       description:""`
	Name        string     `json:"name"        orm:"name"         description:""`
	Phone       string     `json:"phone"       orm:"phone"        description:""`
	EffectAt    *time.Time `json:"effectAt"    orm:"effect_at"    description:""`
	ExpireAt    *time.Time `json:"expireAt"    orm:"expire_at"    description:""`
	GroupId     int64      `json:"groupId"     orm:"group_id"     description:""`
	PassLevel   int        `json:"passLevel"   orm:"pass_level"   description:""`
	ContainerId int64      `json:"containerId" orm:"container_id" description:""`
	UnitId      int64      `json:"unitId"      orm:"unit_id"      description:""`
	Status      int        `json:"status"      orm:"status"       description:""`
	CreatedAt   *time.Time `json:"createdAt"   orm:"created_at"   description:""`
	UpdatedAt   *time.Time `json:"updatedAt"   orm:"updated_at"   description:""`
	DeletedAt   *time.Time `json:"deletedAt"   orm:"deleted_at"   description:""`
	CreateBy    int64      `json:"createBy"    orm:"create_by"    description:""`
	UpdateBy    int64      `json:"updateBy"    orm:"update_by"    description:""`
}
