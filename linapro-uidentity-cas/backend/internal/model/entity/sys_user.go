// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

import (
	"time"
)

// SysUser is the golang structure for table sys_user.
type SysUser struct {
	UserId    int64      `json:"userId"    orm:"user_id"    description:""`
	Username  string     `json:"username"  orm:"username"   description:""`
	Password  string     `json:"password"  orm:"password"   description:""`
	NickName  string     `json:"nickName"  orm:"nick_name"  description:""`
	Phone     string     `json:"phone"     orm:"phone"      description:""`
	RoleId    int64      `json:"roleId"    orm:"role_id"    description:""`
	Salt      string     `json:"salt"      orm:"salt"       description:""`
	Avatar    string     `json:"avatar"    orm:"avatar"     description:""`
	Sex       string     `json:"sex"       orm:"sex"        description:""`
	Email     string     `json:"email"     orm:"email"      description:""`
	DeptId    int64      `json:"deptId"    orm:"dept_id"    description:""`
	PostId    int64      `json:"postId"    orm:"post_id"    description:""`
	Remark    string     `json:"remark"    orm:"remark"     description:""`
	Status    string     `json:"status"    orm:"status"     description:""`
	CreateBy  int64      `json:"createBy"  orm:"create_by"  description:""`
	UpdateBy  int64      `json:"updateBy"  orm:"update_by"  description:""`
	CreatedAt *time.Time `json:"createdAt" orm:"created_at" description:""`
	UpdatedAt *time.Time `json:"updatedAt" orm:"updated_at" description:""`
	DeletedAt *time.Time `json:"deletedAt" orm:"deleted_at" description:""`
}
