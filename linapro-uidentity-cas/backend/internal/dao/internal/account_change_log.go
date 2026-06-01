// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// AccountChangeLogDao is the data access object for the table plugin_linapro_uidentity_cas_account_change_log.
type AccountChangeLogDao struct {
	table    string                  // table is the underlying table name of the DAO.
	group    string                  // group is the database configuration group name of the current DAO.
	columns  AccountChangeLogColumns // columns contains all the column names of Table for convenient usage.
	handlers []gdb.ModelHandler      // handlers for customized model modification.
}

// AccountChangeLogColumns defines and stores column names for the table plugin_linapro_uidentity_cas_account_change_log.
type AccountChangeLogColumns struct {
	Id        string //
	TenantId  string //
	AccountId string //
	TableName string //
	Action    string //
	DataOld   string //
	DataNew   string //
	ErrMsg    string //
	ErrNumber string //
	CreatedBy string //
	UpdatedBy string //
	CreatedAt string //
	UpdatedAt string //
	DeletedAt string //
}

// accountChangeLogColumns holds the columns for the table plugin_linapro_uidentity_cas_account_change_log.
var accountChangeLogColumns = AccountChangeLogColumns{
	Id:        "id",
	TenantId:  "tenant_id",
	AccountId: "account_id",
	TableName: "table_name",
	Action:    "action",
	DataOld:   "data_old",
	DataNew:   "data_new",
	ErrMsg:    "err_msg",
	ErrNumber: "err_number",
	CreatedBy: "created_by",
	UpdatedBy: "updated_by",
	CreatedAt: "created_at",
	UpdatedAt: "updated_at",
	DeletedAt: "deleted_at",
}

// NewAccountChangeLogDao creates and returns a new DAO object for table data access.
func NewAccountChangeLogDao(handlers ...gdb.ModelHandler) *AccountChangeLogDao {
	return &AccountChangeLogDao{
		group:    "default",
		table:    "plugin_linapro_uidentity_cas_account_change_log",
		columns:  accountChangeLogColumns,
		handlers: handlers,
	}
}

// DB retrieves and returns the underlying raw database management object of the current DAO.
func (dao *AccountChangeLogDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of the current DAO.
func (dao *AccountChangeLogDao) Table() string {
	return dao.table
}

// Columns returns all column names of the current DAO.
func (dao *AccountChangeLogDao) Columns() AccountChangeLogColumns {
	return dao.columns
}

// Group returns the database configuration group name of the current DAO.
func (dao *AccountChangeLogDao) Group() string {
	return dao.group
}

// Ctx creates and returns a Model for the current DAO. It automatically sets the context for the current operation.
func (dao *AccountChangeLogDao) Ctx(ctx context.Context) *gdb.Model {
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
func (dao *AccountChangeLogDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
