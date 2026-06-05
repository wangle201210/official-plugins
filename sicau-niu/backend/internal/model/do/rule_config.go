// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"time"

	"github.com/gogf/gf/v2/frame/g"
)

// RuleConfig is the golang structure of table plugin_sicau_niu_rule_config for DAO operations like Where/Data.
type RuleConfig struct {
	g.Meta      `orm:"table:plugin_sicau_niu_rule_config, do:true"`
	Id          any        //
	ConfigKey   any        // Stable runtime rule key
	ConfigValue any        // Runtime rule value stored as text and validated by service
	Remark      any        // Operator-facing description of the rule
	CreatedAt   *time.Time // Creation time
	UpdatedAt   *time.Time // Update time
	DeletedAt   *time.Time // Soft-delete time, NULL means active
}
