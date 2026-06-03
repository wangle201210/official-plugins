// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

import (
	"time"
)

// Steal is the golang structure for table steal.
type Steal struct {
	Id           int64      `json:"id"           orm:"id"             description:""`
	ActorUserId  int64      `json:"actorUserId"  orm:"actor_user_id"  description:""`
	TargetUserId int64      `json:"targetUserId" orm:"target_user_id" description:""`
	Amount       int        `json:"amount"       orm:"amount"         description:""`
	StealDate    string     `json:"stealDate"    orm:"steal_date"     description:"Steal date YYYY-MM-DD for the daily count limit"`
	CreatedAt    *time.Time `json:"createdAt"    orm:"created_at"     description:""`
	UpdatedAt    *time.Time `json:"updatedAt"    orm:"updated_at"     description:""`
	DeletedAt    *time.Time `json:"deletedAt"    orm:"deleted_at"     description:""`
}
