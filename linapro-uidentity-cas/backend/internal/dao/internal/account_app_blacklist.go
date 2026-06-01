// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// AccountAppBlacklistDao is the data access object for the table plugin_linapro_uidentity_cas_account_app_blacklist.
type AccountAppBlacklistDao struct {
	table    string                     // table is the underlying table name of the DAO.
	group    string                     // group is the database configuration group name of the current DAO.
	columns  AccountAppBlacklistColumns // columns contains all the column names of Table for convenient usage.
	handlers []gdb.ModelHandler         // handlers for customized model modification.
}

// AccountAppBlacklistColumns defines and stores column names for the table plugin_linapro_uidentity_cas_account_app_blacklist.
type AccountAppBlacklistColumns struct {
	Id        string //
	TenantId  string //
	Name      string //
	AppId     string //
	AccountId string //
	EffectAt  string //
	ExpireAt  string //
	CreatedBy string //
	UpdatedBy string //
	CreatedAt string //
	UpdatedAt string //
	DeletedAt string //
}

// accountAppBlacklistColumns holds the columns for the table plugin_linapro_uidentity_cas_account_app_blacklist.
var accountAppBlacklistColumns = AccountAppBlacklistColumns{
	Id:        "id",
	TenantId:  "tenant_id",
	Name:      "name",
	AppId:     "app_id",
	AccountId: "account_id",
	EffectAt:  "effect_at",
	ExpireAt:  "expire_at",
	CreatedBy: "created_by",
	UpdatedBy: "updated_by",
	CreatedAt: "created_at",
	UpdatedAt: "updated_at",
	DeletedAt: "deleted_at",
}

// NewAccountAppBlacklistDao creates and returns a new DAO object for table data access.
func NewAccountAppBlacklistDao(handlers ...gdb.ModelHandler) *AccountAppBlacklistDao {
	return &AccountAppBlacklistDao{
		group:    "default",
		table:    "plugin_linapro_uidentity_cas_account_app_blacklist",
		columns:  accountAppBlacklistColumns,
		handlers: handlers,
	}
}

// DB retrieves and returns the underlying raw database management object of the current DAO.
func (dao *AccountAppBlacklistDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of the current DAO.
func (dao *AccountAppBlacklistDao) Table() string {
	return dao.table
}

// Columns returns all column names of the current DAO.
func (dao *AccountAppBlacklistDao) Columns() AccountAppBlacklistColumns {
	return dao.columns
}

// Group returns the database configuration group name of the current DAO.
func (dao *AccountAppBlacklistDao) Group() string {
	return dao.group
}

// Ctx creates and returns a Model for the current DAO. It automatically sets the context for the current operation.
func (dao *AccountAppBlacklistDao) Ctx(ctx context.Context) *gdb.Model {
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
func (dao *AccountAppBlacklistDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
