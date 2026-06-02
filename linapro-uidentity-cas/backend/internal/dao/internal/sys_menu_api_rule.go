// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// SysMenuApiRuleDao is the data access object for the table plugin_linapro_uidentity_cas_sys_menu_api_rule.
type SysMenuApiRuleDao struct {
	table    string                // table is the underlying table name of the DAO.
	group    string                // group is the database configuration group name of the current DAO.
	columns  SysMenuApiRuleColumns // columns contains all the column names of Table for convenient usage.
	handlers []gdb.ModelHandler    // handlers for customized model modification.
}

// SysMenuApiRuleColumns defines and stores column names for the table plugin_linapro_uidentity_cas_sys_menu_api_rule.
type SysMenuApiRuleColumns struct {
	SysMenuMenuId string //
	SysApiId      string //
}

// sysMenuApiRuleColumns holds the columns for the table plugin_linapro_uidentity_cas_sys_menu_api_rule.
var sysMenuApiRuleColumns = SysMenuApiRuleColumns{
	SysMenuMenuId: "sys_menu_menu_id",
	SysApiId:      "sys_api_id",
}

// NewSysMenuApiRuleDao creates and returns a new DAO object for table data access.
func NewSysMenuApiRuleDao(handlers ...gdb.ModelHandler) *SysMenuApiRuleDao {
	return &SysMenuApiRuleDao{
		group:    "default",
		table:    "plugin_linapro_uidentity_cas_sys_menu_api_rule",
		columns:  sysMenuApiRuleColumns,
		handlers: handlers,
	}
}

// DB retrieves and returns the underlying raw database management object of the current DAO.
func (dao *SysMenuApiRuleDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of the current DAO.
func (dao *SysMenuApiRuleDao) Table() string {
	return dao.table
}

// Columns returns all column names of the current DAO.
func (dao *SysMenuApiRuleDao) Columns() SysMenuApiRuleColumns {
	return dao.columns
}

// Group returns the database configuration group name of the current DAO.
func (dao *SysMenuApiRuleDao) Group() string {
	return dao.group
}

// Ctx creates and returns a Model for the current DAO. It automatically sets the context for the current operation.
func (dao *SysMenuApiRuleDao) Ctx(ctx context.Context) *gdb.Model {
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
func (dao *SysMenuApiRuleDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
