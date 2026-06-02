// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// Oauth2TokenDao is the data access object for the table plugin_linapro_uidentity_cas_oauth2_token.
type Oauth2TokenDao struct {
	table    string             // table is the underlying table name of the DAO.
	group    string             // group is the database configuration group name of the current DAO.
	columns  Oauth2TokenColumns // columns contains all the column names of Table for convenient usage.
	handlers []gdb.ModelHandler // handlers for customized model modification.
}

// Oauth2TokenColumns defines and stores column names for the table plugin_linapro_uidentity_cas_oauth2_token.
type Oauth2TokenColumns struct {
	Id        string //
	ExpiredAt string //
	Code      string //
	Access    string //
	Refresh   string //
	Data      string //
	CreatedAt string //
	UpdatedAt string //
	DeletedAt string //
	CreateBy  string //
	UpdateBy  string //
}

// oauth2TokenColumns holds the columns for the table plugin_linapro_uidentity_cas_oauth2_token.
var oauth2TokenColumns = Oauth2TokenColumns{
	Id:        "id",
	ExpiredAt: "expired_at",
	Code:      "code",
	Access:    "access",
	Refresh:   "refresh",
	Data:      "data",
	CreatedAt: "created_at",
	UpdatedAt: "updated_at",
	DeletedAt: "deleted_at",
	CreateBy:  "create_by",
	UpdateBy:  "update_by",
}

// NewOauth2TokenDao creates and returns a new DAO object for table data access.
func NewOauth2TokenDao(handlers ...gdb.ModelHandler) *Oauth2TokenDao {
	return &Oauth2TokenDao{
		group:    "default",
		table:    "plugin_linapro_uidentity_cas_oauth2_token",
		columns:  oauth2TokenColumns,
		handlers: handlers,
	}
}

// DB retrieves and returns the underlying raw database management object of the current DAO.
func (dao *Oauth2TokenDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of the current DAO.
func (dao *Oauth2TokenDao) Table() string {
	return dao.table
}

// Columns returns all column names of the current DAO.
func (dao *Oauth2TokenDao) Columns() Oauth2TokenColumns {
	return dao.columns
}

// Group returns the database configuration group name of the current DAO.
func (dao *Oauth2TokenDao) Group() string {
	return dao.group
}

// Ctx creates and returns a Model for the current DAO. It automatically sets the context for the current operation.
func (dao *Oauth2TokenDao) Ctx(ctx context.Context) *gdb.Model {
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
func (dao *Oauth2TokenDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
