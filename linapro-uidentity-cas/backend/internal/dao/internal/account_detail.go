// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// AccountDetailDao is the data access object for the table plugin_linapro_uidentity_cas_account_detail.
type AccountDetailDao struct {
	table    string               // table is the underlying table name of the DAO.
	group    string               // group is the database configuration group name of the current DAO.
	columns  AccountDetailColumns // columns contains all the column names of Table for convenient usage.
	handlers []gdb.ModelHandler   // handlers for customized model modification.
}

// AccountDetailColumns defines and stores column names for the table plugin_linapro_uidentity_cas_account_detail.
type AccountDetailColumns struct {
	AccountId    string //
	TenantId     string //
	Birthday     string // Date-only birthday in YYYY-MM-DD format
	Email        string //
	Gender       string //
	Qq           string //
	Wechat       string //
	Idcard       string //
	Avatar       string //
	Source       string //
	Grade        string //
	College      string //
	CollegeCode  string //
	Campus       string //
	SchoolSystem string //
	GraduatedAt  string //
	Major        string //
	ClassName    string //
	Face         string //
	CreatedBy    string //
	UpdatedBy    string //
	CreatedAt    string //
	UpdatedAt    string //
}

// accountDetailColumns holds the columns for the table plugin_linapro_uidentity_cas_account_detail.
var accountDetailColumns = AccountDetailColumns{
	AccountId:    "account_id",
	TenantId:     "tenant_id",
	Birthday:     "birthday",
	Email:        "email",
	Gender:       "gender",
	Qq:           "qq",
	Wechat:       "wechat",
	Idcard:       "idcard",
	Avatar:       "avatar",
	Source:       "source",
	Grade:        "grade",
	College:      "college",
	CollegeCode:  "college_code",
	Campus:       "campus",
	SchoolSystem: "school_system",
	GraduatedAt:  "graduated_at",
	Major:        "major",
	ClassName:    "class_name",
	Face:         "face",
	CreatedBy:    "created_by",
	UpdatedBy:    "updated_by",
	CreatedAt:    "created_at",
	UpdatedAt:    "updated_at",
}

// NewAccountDetailDao creates and returns a new DAO object for table data access.
func NewAccountDetailDao(handlers ...gdb.ModelHandler) *AccountDetailDao {
	return &AccountDetailDao{
		group:    "default",
		table:    "plugin_linapro_uidentity_cas_account_detail",
		columns:  accountDetailColumns,
		handlers: handlers,
	}
}

// DB retrieves and returns the underlying raw database management object of the current DAO.
func (dao *AccountDetailDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of the current DAO.
func (dao *AccountDetailDao) Table() string {
	return dao.table
}

// Columns returns all column names of the current DAO.
func (dao *AccountDetailDao) Columns() AccountDetailColumns {
	return dao.columns
}

// Group returns the database configuration group name of the current DAO.
func (dao *AccountDetailDao) Group() string {
	return dao.group
}

// Ctx creates and returns a Model for the current DAO. It automatically sets the context for the current operation.
func (dao *AccountDetailDao) Ctx(ctx context.Context) *gdb.Model {
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
func (dao *AccountDetailDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
