// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

import (
	"time"
)

// Containers is the golang structure for table containers.
type Containers struct {
	Id           int64      `json:"id"           orm:"id"            description:""`
	Name         string     `json:"name"         orm:"name"          description:""`
	Alias        string     `json:"alias"        orm:"alias"         description:""`
	AccountCount int64      `json:"accountCount" orm:"account_count" description:""`
	AdminCount   int64      `json:"adminCount"   orm:"admin_count"   description:""`
	CreatedAt    *time.Time `json:"createdAt"    orm:"created_at"    description:""`
	UpdatedAt    *time.Time `json:"updatedAt"    orm:"updated_at"    description:""`
	DeletedAt    *time.Time `json:"deletedAt"    orm:"deleted_at"    description:""`
	CreateBy     int64      `json:"createBy"     orm:"create_by"     description:""`
	UpdateBy     int64      `json:"updateBy"     orm:"update_by"     description:""`
}
