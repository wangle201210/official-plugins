// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// AccountGroupDao is the data access object for the table plugin_linapro_uidentity_cas_account_group.
type AccountGroupDao struct {
	table    string              // table is the underlying table name of the DAO.
	group    string              // group is the database configuration group name of the current DAO.
	columns  AccountGroupColumns // columns contains all the column names of Table for convenient usage.
	handlers []gdb.ModelHandler  // handlers for customized model modification.
}

// AccountGroupColumns defines and stores column names for the table plugin_linapro_uidentity_cas_account_group.
type AccountGroupColumns struct {
	Id        string //
	TenantId  string //
	AccountId string //
	GroupId   string //
	CreatedBy string //
	CreatedAt string //
	UpdatedAt string //
}

// accountGroupColumns holds the columns for the table plugin_linapro_uidentity_cas_account_group.
var accountGroupColumns = AccountGroupColumns{
	Id:        "id",
	TenantId:  "tenant_id",
	AccountId: "account_id",
	GroupId:   "group_id",
	CreatedBy: "created_by",
	CreatedAt: "created_at",
	UpdatedAt: "updated_at",
}

// NewAccountGroupDao creates and returns a new DAO object for table data access.
func NewAccountGroupDao(handlers ...gdb.ModelHandler) *AccountGroupDao {
	return &AccountGroupDao{
		group:    "default",
		table:    "plugin_linapro_uidentity_cas_account_group",
		columns:  accountGroupColumns,
		handlers: handlers,
	}
}

// DB retrieves and returns the underlying raw database management object of the current DAO.
func (dao *AccountGroupDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of the current DAO.
func (dao *AccountGroupDao) Table() string {
	return dao.table
}

// Columns returns all column names of the current DAO.
func (dao *AccountGroupDao) Columns() AccountGroupColumns {
	return dao.columns
}

// Group returns the database configuration group name of the current DAO.
func (dao *AccountGroupDao) Group() string {
	return dao.group
}

// Ctx creates and returns a Model for the current DAO. It automatically sets the context for the current operation.
func (dao *AccountGroupDao) Ctx(ctx context.Context) *gdb.Model {
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
func (dao *AccountGroupDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
