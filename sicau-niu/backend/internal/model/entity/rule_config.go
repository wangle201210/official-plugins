// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

import (
	"time"
)

// RuleConfig is the golang structure for table rule_config.
type RuleConfig struct {
	Id          int64      `json:"id"          orm:"id"           description:""`
	ConfigKey   string     `json:"configKey"   orm:"config_key"   description:"Stable runtime rule key"`
	ConfigValue string     `json:"configValue" orm:"config_value" description:"Runtime rule value stored as text and validated by service"`
	Remark      string     `json:"remark"      orm:"remark"       description:"Operator-facing description of the rule"`
	CreatedAt   *time.Time `json:"createdAt"   orm:"created_at"   description:"Creation time"`
	UpdatedAt   *time.Time `json:"updatedAt"   orm:"updated_at"   description:"Update time"`
	DeletedAt   *time.Time `json:"deletedAt"   orm:"deleted_at"   description:"Soft-delete time, NULL means active"`
}
