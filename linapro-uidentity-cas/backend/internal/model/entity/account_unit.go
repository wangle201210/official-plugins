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
	AccountId int64      `json:"accountId" orm:"account_id" description:""`
	UnitsId   int64      `json:"unitsId"   orm:"units_id"   description:""`
	CreatedAt *time.Time `json:"createdAt" orm:"created_at" description:""`
	UpdatedAt *time.Time `json:"updatedAt" orm:"updated_at" description:""`
	DeletedAt *time.Time `json:"deletedAt" orm:"deleted_at" description:""`
	CreateBy  int64      `json:"createBy"  orm:"create_by"  description:""`
	UpdateBy  int64      `json:"updateBy"  orm:"update_by"  description:""`
}
