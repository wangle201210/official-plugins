// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// AccountActiveLogDao is the data access object for the table plugin_linapro_uidentity_cas_account_active_log.
type AccountActiveLogDao struct {
	table    string                  // table is the underlying table name of the DAO.
	group    string                  // group is the database configuration group name of the current DAO.
	columns  AccountActiveLogColumns // columns contains all the column names of Table for convenient usage.
	handlers []gdb.ModelHandler      // handlers for customized model modification.
}

// AccountActiveLogColumns defines and stores column names for the table plugin_linapro_uidentity_cas_account_active_log.
type AccountActiveLogColumns struct {
	Id        string //
	TenantId  string //
	Number    string //
	Phone     string //
	Wechat    string //
	Type      string // Legacy activation log type: 0=activation or Wechat rebind callback, 1=union ID bind
	CreatedBy string //
	UpdatedBy string //
	CreatedAt string //
	UpdatedAt string //
	DeletedAt string //
}

// accountActiveLogColumns holds the columns for the table plugin_linapro_uidentity_cas_account_active_log.
var accountActiveLogColumns = AccountActiveLogColumns{
	Id:        "id",
	TenantId:  "tenant_id",
	Number:    "number",
	Phone:     "phone",
	Wechat:    "wechat",
	Type:      "type",
	CreatedBy: "created_by",
	UpdatedBy: "updated_by",
	CreatedAt: "created_at",
	UpdatedAt: "updated_at",
	DeletedAt: "deleted_at",
}

// NewAccountActiveLogDao creates and returns a new DAO object for table data access.
func NewAccountActiveLogDao(handlers ...gdb.ModelHandler) *AccountActiveLogDao {
	return &AccountActiveLogDao{
		group:    "default",
		table:    "plugin_linapro_uidentity_cas_account_active_log",
		columns:  accountActiveLogColumns,
		handlers: handlers,
	}
}

// DB retrieves and returns the underlying raw database management object of the current DAO.
func (dao *AccountActiveLogDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of the current DAO.
func (dao *AccountActiveLogDao) Table() string {
	return dao.table
}

// Columns returns all column names of the current DAO.
func (dao *AccountActiveLogDao) Columns() AccountActiveLogColumns {
	return dao.columns
}

// Group returns the database configuration group name of the current DAO.
func (dao *AccountActiveLogDao) Group() string {
	return dao.group
}

// Ctx creates and returns a Model for the current DAO. It automatically sets the context for the current operation.
func (dao *AccountActiveLogDao) Ctx(ctx context.Context) *gdb.Model {
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
func (dao *AccountActiveLogDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
