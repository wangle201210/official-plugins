// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

import (
	"time"
)

// SysConfig is the golang structure for table sys_config.
type SysConfig struct {
	Id          int64      `json:"id"          orm:"id"           description:""`
	ConfigName  string     `json:"configName"  orm:"config_name"  description:""`
	ConfigKey   string     `json:"configKey"   orm:"config_key"   description:""`
	ConfigValue string     `json:"configValue" orm:"config_value" description:""`
	ConfigType  string     `json:"configType"  orm:"config_type"  description:""`
	IsFrontend  string     `json:"isFrontend"  orm:"is_frontend"  description:""`
	Remark      string     `json:"remark"      orm:"remark"       description:""`
	CreateBy    int64      `json:"createBy"    orm:"create_by"    description:""`
	UpdateBy    int64      `json:"updateBy"    orm:"update_by"    description:""`
	CreatedAt   *time.Time `json:"createdAt"   orm:"created_at"   description:""`
	UpdatedAt   *time.Time `json:"updatedAt"   orm:"updated_at"   description:""`
	DeletedAt   *time.Time `json:"deletedAt"   orm:"deleted_at"   description:""`
}
