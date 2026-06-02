// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

import (
	"time"
)

// SysRole is the golang structure for table sys_role.
type SysRole struct {
	RoleId    int64      `json:"roleId"    orm:"role_id"    description:""`
	RoleName  string     `json:"roleName"  orm:"role_name"  description:""`
	Status    string     `json:"status"    orm:"status"     description:""`
	RoleKey   string     `json:"roleKey"   orm:"role_key"   description:""`
	RoleSort  int64      `json:"roleSort"  orm:"role_sort"  description:""`
	Flag      string     `json:"flag"      orm:"flag"       description:""`
	Remark    string     `json:"remark"    orm:"remark"     description:""`
	Admin     bool       `json:"admin"     orm:"admin"      description:""`
	DataScope string     `json:"dataScope" orm:"data_scope" description:""`
	CreateBy  int64      `json:"createBy"  orm:"create_by"  description:""`
	UpdateBy  int64      `json:"updateBy"  orm:"update_by"  description:""`
	CreatedAt *time.Time `json:"createdAt" orm:"created_at" description:""`
	UpdatedAt *time.Time `json:"updatedAt" orm:"updated_at" description:""`
	DeletedAt *time.Time `json:"deletedAt" orm:"deleted_at" description:""`
}
