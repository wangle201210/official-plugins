// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// GrassAccountDao is the data access object for the table plugin_sicau_niu_grass_account.
type GrassAccountDao struct {
	table    string              // table is the underlying table name of the DAO.
	group    string              // group is the database configuration group name of the current DAO.
	columns  GrassAccountColumns // columns contains all the column names of Table for convenient usage.
	handlers []gdb.ModelHandler  // handlers for customized model modification.
}

// GrassAccountColumns defines and stores column names for the table plugin_sicau_niu_grass_account.
type GrassAccountColumns struct {
	Id        string //
	UserId    string // Owning player ID
	Balance   string // Current grass balance
	CreatedAt string //
	UpdatedAt string //
	DeletedAt string // Soft-delete time, NULL means active
}

// grassAccountColumns holds the columns for the table plugin_sicau_niu_grass_account.
var grassAccountColumns = GrassAccountColumns{
	Id:        "id",
	UserId:    "user_id",
	Balance:   "balance",
	CreatedAt: "created_at",
	UpdatedAt: "updated_at",
	DeletedAt: "deleted_at",
}

// NewGrassAccountDao creates and returns a new DAO object for table data access.
func NewGrassAccountDao(handlers ...gdb.ModelHandler) *GrassAccountDao {
	return &GrassAccountDao{
		group:    "default",
		table:    "plugin_sicau_niu_grass_account",
		columns:  grassAccountColumns,
		handlers: handlers,
	}
}

// DB retrieves and returns the underlying raw database management object of the current DAO.
func (dao *GrassAccountDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of the current DAO.
func (dao *GrassAccountDao) Table() string {
	return dao.table
}

// Columns returns all column names of the current DAO.
func (dao *GrassAccountDao) Columns() GrassAccountColumns {
	return dao.columns
}

// Group returns the database configuration group name of the current DAO.
func (dao *GrassAccountDao) Group() string {
	return dao.group
}

// Ctx creates and returns a Model for the current DAO. It automatically sets the context for the current operation.
func (dao *GrassAccountDao) Ctx(ctx context.Context) *gdb.Model {
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
func (dao *GrassAccountDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
