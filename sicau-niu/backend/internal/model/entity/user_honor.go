// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

import (
	"time"
)

// UserHonor is the golang structure for table user_honor.
type UserHonor struct {
	Id         int64      `json:"id"         orm:"id"          description:""`
	UserId     int64      `json:"userId"     orm:"user_id"     description:"Player ID"`
	HonorId    int64      `json:"honorId"    orm:"honor_id"    description:"Granted honor definition ID"`
	UnlockedAt *time.Time `json:"unlockedAt" orm:"unlocked_at" description:"Grant time"`
	CreatedAt  *time.Time `json:"createdAt"  orm:"created_at"  description:""`
	UpdatedAt  *time.Time `json:"updatedAt"  orm:"updated_at"  description:""`
	DeletedAt  *time.Time `json:"deletedAt"  orm:"deleted_at"  description:""`
}
