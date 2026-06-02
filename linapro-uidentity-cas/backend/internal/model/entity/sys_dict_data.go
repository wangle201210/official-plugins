// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

import (
	"time"
)

// SysDictData is the golang structure for table sys_dict_data.
type SysDictData struct {
	DictCode  int64      `json:"dictCode"  orm:"dict_code"  description:""`
	DictSort  int64      `json:"dictSort"  orm:"dict_sort"  description:""`
	DictLabel string     `json:"dictLabel" orm:"dict_label" description:""`
	DictValue string     `json:"dictValue" orm:"dict_value" description:""`
	DictType  string     `json:"dictType"  orm:"dict_type"  description:""`
	CssClass  string     `json:"cssClass"  orm:"css_class"  description:""`
	ListClass string     `json:"listClass" orm:"list_class" description:""`
	IsDefault string     `json:"isDefault" orm:"is_default" description:""`
	Status    int64      `json:"status"    orm:"status"     description:""`
	Default   string     `json:"default"   orm:"default"    description:""`
	Remark    string     `json:"remark"    orm:"remark"     description:""`
	CreateBy  int64      `json:"createBy"  orm:"create_by"  description:""`
	UpdateBy  int64      `json:"updateBy"  orm:"update_by"  description:""`
	CreatedAt *time.Time `json:"createdAt" orm:"created_at" description:""`
	UpdatedAt *time.Time `json:"updatedAt" orm:"updated_at" description:""`
	DeletedAt *time.Time `json:"deletedAt" orm:"deleted_at" description:""`
}
