// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

import (
	"time"
)

// SysDept is the golang structure for table sys_dept.
type SysDept struct {
	DeptId    int64      `json:"deptId"    orm:"dept_id"    description:""`
	ParentId  int64      `json:"parentId"  orm:"parent_id"  description:""`
	DeptPath  string     `json:"deptPath"  orm:"dept_path"  description:""`
	DeptName  string     `json:"deptName"  orm:"dept_name"  description:""`
	Sort      int64      `json:"sort"      orm:"sort"       description:""`
	Leader    string     `json:"leader"    orm:"leader"     description:""`
	Phone     string     `json:"phone"     orm:"phone"      description:""`
	Email     string     `json:"email"     orm:"email"      description:""`
	Status    int64      `json:"status"    orm:"status"     description:""`
	CreateBy  int64      `json:"createBy"  orm:"create_by"  description:""`
	UpdateBy  int64      `json:"updateBy"  orm:"update_by"  description:""`
	CreatedAt *time.Time `json:"createdAt" orm:"created_at" description:""`
	UpdatedAt *time.Time `json:"updatedAt" orm:"updated_at" description:""`
	DeletedAt *time.Time `json:"deletedAt" orm:"deleted_at" description:""`
}
