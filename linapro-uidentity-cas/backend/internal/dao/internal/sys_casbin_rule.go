// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// SysCasbinRuleDao is the data access object for the table plugin_linapro_uidentity_cas_sys_casbin_rule.
type SysCasbinRuleDao struct {
	table    string               // table is the underlying table name of the DAO.
	group    string               // group is the database configuration group name of the current DAO.
	columns  SysCasbinRuleColumns // columns contains all the column names of Table for convenient usage.
	handlers []gdb.ModelHandler   // handlers for customized model modification.
}

// SysCasbinRuleColumns defines and stores column names for the table plugin_linapro_uidentity_cas_sys_casbin_rule.
type SysCasbinRuleColumns struct {
	Id    string //
	Ptype string //
	V0    string //
	V1    string //
	V2    string //
	V3    string //
	V4    string //
	V5    string //
}

// sysCasbinRuleColumns holds the columns for the table plugin_linapro_uidentity_cas_sys_casbin_rule.
var sysCasbinRuleColumns = SysCasbinRuleColumns{
	Id:    "id",
	Ptype: "ptype",
	V0:    "v0",
	V1:    "v1",
	V2:    "v2",
	V3:    "v3",
	V4:    "v4",
	V5:    "v5",
}

// NewSysCasbinRuleDao creates and returns a new DAO object for table data access.
func NewSysCasbinRuleDao(handlers ...gdb.ModelHandler) *SysCasbinRuleDao {
	return &SysCasbinRuleDao{
		group:    "default",
		table:    "plugin_linapro_uidentity_cas_sys_casbin_rule",
		columns:  sysCasbinRuleColumns,
		handlers: handlers,
	}
}

// DB retrieves and returns the underlying raw database management object of the current DAO.
func (dao *SysCasbinRuleDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of the current DAO.
func (dao *SysCasbinRuleDao) Table() string {
	return dao.table
}

// Columns returns all column names of the current DAO.
func (dao *SysCasbinRuleDao) Columns() SysCasbinRuleColumns {
	return dao.columns
}

// Group returns the database configuration group name of the current DAO.
func (dao *SysCasbinRuleDao) Group() string {
	return dao.group
}

// Ctx creates and returns a Model for the current DAO. It automatically sets the context for the current operation.
func (dao *SysCasbinRuleDao) Ctx(ctx context.Context) *gdb.Model {
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
func (dao *SysCasbinRuleDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
