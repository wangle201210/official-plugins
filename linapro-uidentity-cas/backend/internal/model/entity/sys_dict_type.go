// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

import (
	"time"
)

// SysDictType is the golang structure for table sys_dict_type.
type SysDictType struct {
	DictId    int64      `json:"dictId"    orm:"dict_id"    description:""`
	DictName  string     `json:"dictName"  orm:"dict_name"  description:""`
	DictType  string     `json:"dictType"  orm:"dict_type"  description:""`
	Status    int64      `json:"status"    orm:"status"     description:""`
	Remark    string     `json:"remark"    orm:"remark"     description:""`
	CreateBy  int64      `json:"createBy"  orm:"create_by"  description:""`
	UpdateBy  int64      `json:"updateBy"  orm:"update_by"  description:""`
	CreatedAt *time.Time `json:"createdAt" orm:"created_at" description:""`
	UpdatedAt *time.Time `json:"updatedAt" orm:"updated_at" description:""`
	DeletedAt *time.Time `json:"deletedAt" orm:"deleted_at" description:""`
}
