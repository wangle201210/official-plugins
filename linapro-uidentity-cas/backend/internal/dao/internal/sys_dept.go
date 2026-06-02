// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// SysDeptDao is the data access object for the table plugin_linapro_uidentity_cas_sys_dept.
type SysDeptDao struct {
	table    string             // table is the underlying table name of the DAO.
	group    string             // group is the database configuration group name of the current DAO.
	columns  SysDeptColumns     // columns contains all the column names of Table for convenient usage.
	handlers []gdb.ModelHandler // handlers for customized model modification.
}

// SysDeptColumns defines and stores column names for the table plugin_linapro_uidentity_cas_sys_dept.
type SysDeptColumns struct {
	DeptId    string //
	ParentId  string //
	DeptPath  string //
	DeptName  string //
	Sort      string //
	Leader    string //
	Phone     string //
	Email     string //
	Status    string //
	CreateBy  string //
	UpdateBy  string //
	CreatedAt string //
	UpdatedAt string //
	DeletedAt string //
}

// sysDeptColumns holds the columns for the table plugin_linapro_uidentity_cas_sys_dept.
var sysDeptColumns = SysDeptColumns{
	DeptId:    "dept_id",
	ParentId:  "parent_id",
	DeptPath:  "dept_path",
	DeptName:  "dept_name",
	Sort:      "sort",
	Leader:    "leader",
	Phone:     "phone",
	Email:     "email",
	Status:    "status",
	CreateBy:  "create_by",
	UpdateBy:  "update_by",
	CreatedAt: "created_at",
	UpdatedAt: "updated_at",
	DeletedAt: "deleted_at",
}

// NewSysDeptDao creates and returns a new DAO object for table data access.
func NewSysDeptDao(handlers ...gdb.ModelHandler) *SysDeptDao {
	return &SysDeptDao{
		group:    "default",
		table:    "plugin_linapro_uidentity_cas_sys_dept",
		columns:  sysDeptColumns,
		handlers: handlers,
	}
}

// DB retrieves and returns the underlying raw database management object of the current DAO.
func (dao *SysDeptDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of the current DAO.
func (dao *SysDeptDao) Table() string {
	return dao.table
}

// Columns returns all column names of the current DAO.
func (dao *SysDeptDao) Columns() SysDeptColumns {
	return dao.columns
}

// Group returns the database configuration group name of the current DAO.
func (dao *SysDeptDao) Group() string {
	return dao.group
}

// Ctx creates and returns a Model for the current DAO. It automatically sets the context for the current operation.
func (dao *SysDeptDao) Ctx(ctx context.Context) *gdb.Model {
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
func (dao *SysDeptDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
