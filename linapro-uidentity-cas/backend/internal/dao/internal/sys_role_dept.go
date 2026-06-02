// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// SysRoleDeptDao is the data access object for the table plugin_linapro_uidentity_cas_sys_role_dept.
type SysRoleDeptDao struct {
	table    string             // table is the underlying table name of the DAO.
	group    string             // group is the database configuration group name of the current DAO.
	columns  SysRoleDeptColumns // columns contains all the column names of Table for convenient usage.
	handlers []gdb.ModelHandler // handlers for customized model modification.
}

// SysRoleDeptColumns defines and stores column names for the table plugin_linapro_uidentity_cas_sys_role_dept.
type SysRoleDeptColumns struct {
	RoleId string //
	DeptId string //
}

// sysRoleDeptColumns holds the columns for the table plugin_linapro_uidentity_cas_sys_role_dept.
var sysRoleDeptColumns = SysRoleDeptColumns{
	RoleId: "role_id",
	DeptId: "dept_id",
}

// NewSysRoleDeptDao creates and returns a new DAO object for table data access.
func NewSysRoleDeptDao(handlers ...gdb.ModelHandler) *SysRoleDeptDao {
	return &SysRoleDeptDao{
		group:    "default",
		table:    "plugin_linapro_uidentity_cas_sys_role_dept",
		columns:  sysRoleDeptColumns,
		handlers: handlers,
	}
}

// DB retrieves and returns the underlying raw database management object of the current DAO.
func (dao *SysRoleDeptDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of the current DAO.
func (dao *SysRoleDeptDao) Table() string {
	return dao.table
}

// Columns returns all column names of the current DAO.
func (dao *SysRoleDeptDao) Columns() SysRoleDeptColumns {
	return dao.columns
}

// Group returns the database configuration group name of the current DAO.
func (dao *SysRoleDeptDao) Group() string {
	return dao.group
}

// Ctx creates and returns a Model for the current DAO. It automatically sets the context for the current operation.
func (dao *SysRoleDeptDao) Ctx(ctx context.Context) *gdb.Model {
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
func (dao *SysRoleDeptDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
