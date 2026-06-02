// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

import (
	"time"
)

// SysMenu is the golang structure for table sys_menu.
type SysMenu struct {
	MenuId     int64      `json:"menuId"     orm:"menu_id"    description:""`
	MenuName   string     `json:"menuName"   orm:"menu_name"  description:""`
	Title      string     `json:"title"      orm:"title"      description:""`
	Icon       string     `json:"icon"       orm:"icon"       description:""`
	Path       string     `json:"path"       orm:"path"       description:""`
	Paths      string     `json:"paths"      orm:"paths"      description:""`
	MenuType   string     `json:"menuType"   orm:"menu_type"  description:""`
	Action     string     `json:"action"     orm:"action"     description:""`
	Permission string     `json:"permission" orm:"permission" description:""`
	ParentId   int64      `json:"parentId"   orm:"parent_id"  description:""`
	NoCache    bool       `json:"noCache"    orm:"no_cache"   description:""`
	Breadcrumb string     `json:"breadcrumb" orm:"breadcrumb" description:""`
	Component  string     `json:"component"  orm:"component"  description:""`
	Sort       int64      `json:"sort"       orm:"sort"       description:""`
	Visible    string     `json:"visible"    orm:"visible"    description:""`
	IsFrame    string     `json:"isFrame"    orm:"is_frame"   description:""`
	CreateBy   int64      `json:"createBy"   orm:"create_by"  description:""`
	UpdateBy   int64      `json:"updateBy"   orm:"update_by"  description:""`
	CreatedAt  *time.Time `json:"createdAt"  orm:"created_at" description:""`
	UpdatedAt  *time.Time `json:"updatedAt"  orm:"updated_at" description:""`
	DeletedAt  *time.Time `json:"deletedAt"  orm:"deleted_at" description:""`
}
