// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

import (
	"time"
)

// SysApi is the golang structure for table sys_api.
type SysApi struct {
	Id        int64      `json:"id"        orm:"id"         description:""`
	Handle    string     `json:"handle"    orm:"handle"     description:""`
	Title     string     `json:"title"     orm:"title"      description:""`
	Path      string     `json:"path"      orm:"path"       description:""`
	Type      string     `json:"type"      orm:"type"       description:""`
	Action    string     `json:"action"    orm:"action"     description:""`
	CreateBy  int64      `json:"createBy"  orm:"create_by"  description:""`
	UpdateBy  int64      `json:"updateBy"  orm:"update_by"  description:""`
	CreatedAt *time.Time `json:"createdAt" orm:"created_at" description:""`
	UpdatedAt *time.Time `json:"updatedAt" orm:"updated_at" description:""`
	DeletedAt *time.Time `json:"deletedAt" orm:"deleted_at" description:""`
}
