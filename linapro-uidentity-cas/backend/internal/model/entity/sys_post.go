// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

import (
	"time"
)

// SysPost is the golang structure for table sys_post.
type SysPost struct {
	PostId    int64      `json:"postId"    orm:"post_id"    description:""`
	PostName  string     `json:"postName"  orm:"post_name"  description:""`
	PostCode  string     `json:"postCode"  orm:"post_code"  description:""`
	Sort      int64      `json:"sort"      orm:"sort"       description:""`
	Status    int64      `json:"status"    orm:"status"     description:""`
	Remark    string     `json:"remark"    orm:"remark"     description:""`
	CreateBy  int64      `json:"createBy"  orm:"create_by"  description:""`
	UpdateBy  int64      `json:"updateBy"  orm:"update_by"  description:""`
	CreatedAt *time.Time `json:"createdAt" orm:"created_at" description:""`
	UpdatedAt *time.Time `json:"updatedAt" orm:"updated_at" description:""`
	DeletedAt *time.Time `json:"deletedAt" orm:"deleted_at" description:""`
}
