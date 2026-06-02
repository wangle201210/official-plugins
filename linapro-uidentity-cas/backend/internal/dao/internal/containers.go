// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// ContainersDao is the data access object for the table plugin_linapro_uidentity_cas_containers.
type ContainersDao struct {
	table    string             // table is the underlying table name of the DAO.
	group    string             // group is the database configuration group name of the current DAO.
	columns  ContainersColumns  // columns contains all the column names of Table for convenient usage.
	handlers []gdb.ModelHandler // handlers for customized model modification.
}

// ContainersColumns defines and stores column names for the table plugin_linapro_uidentity_cas_containers.
type ContainersColumns struct {
	Id           string //
	Name         string //
	Alias        string //
	AccountCount string //
	AdminCount   string //
	CreatedAt    string //
	UpdatedAt    string //
	DeletedAt    string //
	CreateBy     string //
	UpdateBy     string //
}

// containersColumns holds the columns for the table plugin_linapro_uidentity_cas_containers.
var containersColumns = ContainersColumns{
	Id:           "id",
	Name:         "name",
	Alias:        "alias",
	AccountCount: "account_count",
	AdminCount:   "admin_count",
	CreatedAt:    "created_at",
	UpdatedAt:    "updated_at",
	DeletedAt:    "deleted_at",
	CreateBy:     "create_by",
	UpdateBy:     "update_by",
}

// NewContainersDao creates and returns a new DAO object for table data access.
func NewContainersDao(handlers ...gdb.ModelHandler) *ContainersDao {
	return &ContainersDao{
		group:    "default",
		table:    "plugin_linapro_uidentity_cas_containers",
		columns:  containersColumns,
		handlers: handlers,
	}
}

// DB retrieves and returns the underlying raw database management object of the current DAO.
func (dao *ContainersDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of the current DAO.
func (dao *ContainersDao) Table() string {
	return dao.table
}

// Columns returns all column names of the current DAO.
func (dao *ContainersDao) Columns() ContainersColumns {
	return dao.columns
}

// Group returns the database configuration group name of the current DAO.
func (dao *ContainersDao) Group() string {
	return dao.group
}

// Ctx creates and returns a Model for the current DAO. It automatically sets the context for the current operation.
func (dao *ContainersDao) Ctx(ctx context.Context) *gdb.Model {
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
func (dao *ContainersDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
