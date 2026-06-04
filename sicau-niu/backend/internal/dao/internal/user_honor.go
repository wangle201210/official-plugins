// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// UserHonorDao is the data access object for the table plugin_sicau_niu_user_honor.
type UserHonorDao struct {
	table    string             // table is the underlying table name of the DAO.
	group    string             // group is the database configuration group name of the current DAO.
	columns  UserHonorColumns   // columns contains all the column names of Table for convenient usage.
	handlers []gdb.ModelHandler // handlers for customized model modification.
}

// UserHonorColumns defines and stores column names for the table plugin_sicau_niu_user_honor.
type UserHonorColumns struct {
	Id         string //
	UserId     string // Player ID
	HonorId    string // Granted honor definition ID
	UnlockedAt string // Grant time
	CreatedAt  string //
	UpdatedAt  string //
	DeletedAt  string //
}

// userHonorColumns holds the columns for the table plugin_sicau_niu_user_honor.
var userHonorColumns = UserHonorColumns{
	Id:         "id",
	UserId:     "user_id",
	HonorId:    "honor_id",
	UnlockedAt: "unlocked_at",
	CreatedAt:  "created_at",
	UpdatedAt:  "updated_at",
	DeletedAt:  "deleted_at",
}

// NewUserHonorDao creates and returns a new DAO object for table data access.
func NewUserHonorDao(handlers ...gdb.ModelHandler) *UserHonorDao {
	return &UserHonorDao{
		group:    "default",
		table:    "plugin_sicau_niu_user_honor",
		columns:  userHonorColumns,
		handlers: handlers,
	}
}

// DB retrieves and returns the underlying raw database management object of the current DAO.
func (dao *UserHonorDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of the current DAO.
func (dao *UserHonorDao) Table() string {
	return dao.table
}

// Columns returns all column names of the current DAO.
func (dao *UserHonorDao) Columns() UserHonorColumns {
	return dao.columns
}

// Group returns the database configuration group name of the current DAO.
func (dao *UserHonorDao) Group() string {
	return dao.group
}

// Ctx creates and returns a Model for the current DAO. It automatically sets the context for the current operation.
func (dao *UserHonorDao) Ctx(ctx context.Context) *gdb.Model {
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
func (dao *UserHonorDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
