// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

import (
	"time"
)

// GrassAccount is the golang structure for table grass_account.
type GrassAccount struct {
	Id        int64      `json:"id"        orm:"id"         description:""`
	UserId    int64      `json:"userId"    orm:"user_id"    description:"Owning player ID"`
	Balance   int64      `json:"balance"   orm:"balance"    description:"Current grass balance"`
	CreatedAt *time.Time `json:"createdAt" orm:"created_at" description:""`
	UpdatedAt *time.Time `json:"updatedAt" orm:"updated_at" description:""`
	DeletedAt *time.Time `json:"deletedAt" orm:"deleted_at" description:"Soft-delete time, NULL means active"`
}
