// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// AccountUnitDao is the data access object for the table plugin_linapro_uidentity_cas_account_unit.
type AccountUnitDao struct {
	table    string             // table is the underlying table name of the DAO.
	group    string             // group is the database configuration group name of the current DAO.
	columns  AccountUnitColumns // columns contains all the column names of Table for convenient usage.
	handlers []gdb.ModelHandler // handlers for customized model modification.
}

// AccountUnitColumns defines and stores column names for the table plugin_linapro_uidentity_cas_account_unit.
type AccountUnitColumns struct {
	Id        string //
	TenantId  string //
	AccountId string //
	UnitId    string //
	CreatedBy string //
	CreatedAt string //
	UpdatedAt string //
}

// accountUnitColumns holds the columns for the table plugin_linapro_uidentity_cas_account_unit.
var accountUnitColumns = AccountUnitColumns{
	Id:        "id",
	TenantId:  "tenant_id",
	AccountId: "account_id",
	UnitId:    "unit_id",
	CreatedBy: "created_by",
	CreatedAt: "created_at",
	UpdatedAt: "updated_at",
}

// NewAccountUnitDao creates and returns a new DAO object for table data access.
func NewAccountUnitDao(handlers ...gdb.ModelHandler) *AccountUnitDao {
	return &AccountUnitDao{
		group:    "default",
		table:    "plugin_linapro_uidentity_cas_account_unit",
		columns:  accountUnitColumns,
		handlers: handlers,
	}
}

// DB retrieves and returns the underlying raw database management object of the current DAO.
func (dao *AccountUnitDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of the current DAO.
func (dao *AccountUnitDao) Table() string {
	return dao.table
}

// Columns returns all column names of the current DAO.
func (dao *AccountUnitDao) Columns() AccountUnitColumns {
	return dao.columns
}

// Group returns the database configuration group name of the current DAO.
func (dao *AccountUnitDao) Group() string {
	return dao.group
}

// Ctx creates and returns a Model for the current DAO. It automatically sets the context for the current operation.
func (dao *AccountUnitDao) Ctx(ctx context.Context) *gdb.Model {
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
func (dao *AccountUnitDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
