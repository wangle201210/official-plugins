// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// AccountDetailsDao is the data access object for the table plugin_linapro_uidentity_cas_account_details.
type AccountDetailsDao struct {
	table    string                // table is the underlying table name of the DAO.
	group    string                // group is the database configuration group name of the current DAO.
	columns  AccountDetailsColumns // columns contains all the column names of Table for convenient usage.
	handlers []gdb.ModelHandler    // handlers for customized model modification.
}

// AccountDetailsColumns defines and stores column names for the table plugin_linapro_uidentity_cas_account_details.
type AccountDetailsColumns struct {
	AccountId string //
	Birthday  string //
	Email     string //
	Gender    string //
	Qq        string //
	Wechat    string //
	Idcard    string //
	Avatar    string //
	Source    string //
	Nj        string //
	Xymc      string //
	Xydm      string //
	Xq        string //
	Xz        string //
	Yjbysj    string //
	Zymc      string //
	Bjmc      string //
	Face      string //
	CreatedAt string //
	UpdatedAt string //
	DeletedAt string //
	CreateBy  string //
	UpdateBy  string //
}

// accountDetailsColumns holds the columns for the table plugin_linapro_uidentity_cas_account_details.
var accountDetailsColumns = AccountDetailsColumns{
	AccountId: "account_id",
	Birthday:  "birthday",
	Email:     "email",
	Gender:    "gender",
	Qq:        "qq",
	Wechat:    "wechat",
	Idcard:    "idcard",
	Avatar:    "avatar",
	Source:    "source",
	Nj:        "nj",
	Xymc:      "xymc",
	Xydm:      "xydm",
	Xq:        "xq",
	Xz:        "xz",
	Yjbysj:    "yjbysj",
	Zymc:      "zymc",
	Bjmc:      "bjmc",
	Face:      "face",
	CreatedAt: "created_at",
	UpdatedAt: "updated_at",
	DeletedAt: "deleted_at",
	CreateBy:  "create_by",
	UpdateBy:  "update_by",
}

// NewAccountDetailsDao creates and returns a new DAO object for table data access.
func NewAccountDetailsDao(handlers ...gdb.ModelHandler) *AccountDetailsDao {
	return &AccountDetailsDao{
		group:    "default",
		table:    "plugin_linapro_uidentity_cas_account_details",
		columns:  accountDetailsColumns,
		handlers: handlers,
	}
}

// DB retrieves and returns the underlying raw database management object of the current DAO.
func (dao *AccountDetailsDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of the current DAO.
func (dao *AccountDetailsDao) Table() string {
	return dao.table
}

// Columns returns all column names of the current DAO.
func (dao *AccountDetailsDao) Columns() AccountDetailsColumns {
	return dao.columns
}

// Group returns the database configuration group name of the current DAO.
func (dao *AccountDetailsDao) Group() string {
	return dao.group
}

// Ctx creates and returns a Model for the current DAO. It automatically sets the context for the current operation.
func (dao *AccountDetailsDao) Ctx(ctx context.Context) *gdb.Model {
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
func (dao *AccountDetailsDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
