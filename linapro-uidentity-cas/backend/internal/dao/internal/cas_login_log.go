// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// CasLoginLogDao is the data access object for the table plugin_linapro_uidentity_cas_cas_login_log.
type CasLoginLogDao struct {
	table    string             // table is the underlying table name of the DAO.
	group    string             // group is the database configuration group name of the current DAO.
	columns  CasLoginLogColumns // columns contains all the column names of Table for convenient usage.
	handlers []gdb.ModelHandler // handlers for customized model modification.
}

// CasLoginLogColumns defines and stores column names for the table plugin_linapro_uidentity_cas_cas_login_log.
type CasLoginLogColumns struct {
	Id              string //
	TenantId        string //
	AccountId       string //
	ChoiceAccountId string //
	AppId           string //
	Ipaddr          string //
	LoginLocation   string //
	Browser         string //
	Os              string //
	Platform        string //
	LoginTime       string //
	Remark          string //
	Msg             string //
	LoginType       string //
	CreatedBy       string //
	UpdatedBy       string //
	CreatedAt       string //
	UpdatedAt       string //
	DeletedAt       string //
}

// casLoginLogColumns holds the columns for the table plugin_linapro_uidentity_cas_cas_login_log.
var casLoginLogColumns = CasLoginLogColumns{
	Id:              "id",
	TenantId:        "tenant_id",
	AccountId:       "account_id",
	ChoiceAccountId: "choice_account_id",
	AppId:           "app_id",
	Ipaddr:          "ipaddr",
	LoginLocation:   "login_location",
	Browser:         "browser",
	Os:              "os",
	Platform:        "platform",
	LoginTime:       "login_time",
	Remark:          "remark",
	Msg:             "msg",
	LoginType:       "login_type",
	CreatedBy:       "created_by",
	UpdatedBy:       "updated_by",
	CreatedAt:       "created_at",
	UpdatedAt:       "updated_at",
	DeletedAt:       "deleted_at",
}

// NewCasLoginLogDao creates and returns a new DAO object for table data access.
func NewCasLoginLogDao(handlers ...gdb.ModelHandler) *CasLoginLogDao {
	return &CasLoginLogDao{
		group:    "default",
		table:    "plugin_linapro_uidentity_cas_cas_login_log",
		columns:  casLoginLogColumns,
		handlers: handlers,
	}
}

// DB retrieves and returns the underlying raw database management object of the current DAO.
func (dao *CasLoginLogDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of the current DAO.
func (dao *CasLoginLogDao) Table() string {
	return dao.table
}

// Columns returns all column names of the current DAO.
func (dao *CasLoginLogDao) Columns() CasLoginLogColumns {
	return dao.columns
}

// Group returns the database configuration group name of the current DAO.
func (dao *CasLoginLogDao) Group() string {
	return dao.group
}

// Ctx creates and returns a Model for the current DAO. It automatically sets the context for the current operation.
func (dao *CasLoginLogDao) Ctx(ctx context.Context) *gdb.Model {
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
func (dao *CasLoginLogDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
