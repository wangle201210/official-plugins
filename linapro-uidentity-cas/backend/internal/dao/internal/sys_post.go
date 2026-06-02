// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// SysPostDao is the data access object for the table plugin_linapro_uidentity_cas_sys_post.
type SysPostDao struct {
	table    string             // table is the underlying table name of the DAO.
	group    string             // group is the database configuration group name of the current DAO.
	columns  SysPostColumns     // columns contains all the column names of Table for convenient usage.
	handlers []gdb.ModelHandler // handlers for customized model modification.
}

// SysPostColumns defines and stores column names for the table plugin_linapro_uidentity_cas_sys_post.
type SysPostColumns struct {
	PostId    string //
	PostName  string //
	PostCode  string //
	Sort      string //
	Status    string //
	Remark    string //
	CreateBy  string //
	UpdateBy  string //
	CreatedAt string //
	UpdatedAt string //
	DeletedAt string //
}

// sysPostColumns holds the columns for the table plugin_linapro_uidentity_cas_sys_post.
var sysPostColumns = SysPostColumns{
	PostId:    "post_id",
	PostName:  "post_name",
	PostCode:  "post_code",
	Sort:      "sort",
	Status:    "status",
	Remark:    "remark",
	CreateBy:  "create_by",
	UpdateBy:  "update_by",
	CreatedAt: "created_at",
	UpdatedAt: "updated_at",
	DeletedAt: "deleted_at",
}

// NewSysPostDao creates and returns a new DAO object for table data access.
func NewSysPostDao(handlers ...gdb.ModelHandler) *SysPostDao {
	return &SysPostDao{
		group:    "default",
		table:    "plugin_linapro_uidentity_cas_sys_post",
		columns:  sysPostColumns,
		handlers: handlers,
	}
}

// DB retrieves and returns the underlying raw database management object of the current DAO.
func (dao *SysPostDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of the current DAO.
func (dao *SysPostDao) Table() string {
	return dao.table
}

// Columns returns all column names of the current DAO.
func (dao *SysPostDao) Columns() SysPostColumns {
	return dao.columns
}

// Group returns the database configuration group name of the current DAO.
func (dao *SysPostDao) Group() string {
	return dao.group
}

// Ctx creates and returns a Model for the current DAO. It automatically sets the context for the current operation.
func (dao *SysPostDao) Ctx(ctx context.Context) *gdb.Model {
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
func (dao *SysPostDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
