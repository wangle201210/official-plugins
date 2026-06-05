// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// RuleConfigDao is the data access object for the table plugin_sicau_niu_rule_config.
type RuleConfigDao struct {
	table    string             // table is the underlying table name of the DAO.
	group    string             // group is the database configuration group name of the current DAO.
	columns  RuleConfigColumns  // columns contains all the column names of Table for convenient usage.
	handlers []gdb.ModelHandler // handlers for customized model modification.
}

// RuleConfigColumns defines and stores column names for the table plugin_sicau_niu_rule_config.
type RuleConfigColumns struct {
	Id          string //
	ConfigKey   string // Stable runtime rule key
	ConfigValue string // Runtime rule value stored as text and validated by service
	Remark      string // Operator-facing description of the rule
	CreatedAt   string // Creation time
	UpdatedAt   string // Update time
	DeletedAt   string // Soft-delete time, NULL means active
}

// ruleConfigColumns holds the columns for the table plugin_sicau_niu_rule_config.
var ruleConfigColumns = RuleConfigColumns{
	Id:          "id",
	ConfigKey:   "config_key",
	ConfigValue: "config_value",
	Remark:      "remark",
	CreatedAt:   "created_at",
	UpdatedAt:   "updated_at",
	DeletedAt:   "deleted_at",
}

// NewRuleConfigDao creates and returns a new DAO object for table data access.
func NewRuleConfigDao(handlers ...gdb.ModelHandler) *RuleConfigDao {
	return &RuleConfigDao{
		group:    "default",
		table:    "plugin_sicau_niu_rule_config",
		columns:  ruleConfigColumns,
		handlers: handlers,
	}
}

// DB retrieves and returns the underlying raw database management object of the current DAO.
func (dao *RuleConfigDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of the current DAO.
func (dao *RuleConfigDao) Table() string {
	return dao.table
}

// Columns returns all column names of the current DAO.
func (dao *RuleConfigDao) Columns() RuleConfigColumns {
	return dao.columns
}

// Group returns the database configuration group name of the current DAO.
func (dao *RuleConfigDao) Group() string {
	return dao.group
}

// Ctx creates and returns a Model for the current DAO. It automatically sets the context for the current operation.
func (dao *RuleConfigDao) Ctx(ctx context.Context) *gdb.Model {
	model := dao.DB().Model(dao.table)
	for _, handler := range dao.handlers {
		model = handler(model)
	}
	return model.Safe().Ctx(ctx)
}

// Transaction wraps the transaction logic using function f.
// It rolls back the transaction and returns the error if function f returns a non-nil error.
// It commits the transaction and returns nil if function f returns nil.
//
// Note: Do not commit or roll back the transaction in function f,
// as it is automatically handled by this function.
func (dao *RuleConfigDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
