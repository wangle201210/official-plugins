// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// ApplicationsDao is the data access object for the table plugin_linapro_uidentity_cas_applications.
type ApplicationsDao struct {
	table    string              // table is the underlying table name of the DAO.
	group    string              // group is the database configuration group name of the current DAO.
	columns  ApplicationsColumns // columns contains all the column names of Table for convenient usage.
	handlers []gdb.ModelHandler  // handlers for customized model modification.
}

// ApplicationsColumns defines and stores column names for the table plugin_linapro_uidentity_cas_applications.
type ApplicationsColumns struct {
	Id          string //
	Name        string //
	Alias       string //
	ClientId    string //
	SecretKey   string //
	AccessModel string //
	Status      string //
	CallbackUrl string //
	Whitelist   string //
	CreatedAt   string //
	UpdatedAt   string //
	DeletedAt   string //
	CreateBy    string //
	UpdateBy    string //
}

// applicationsColumns holds the columns for the table plugin_linapro_uidentity_cas_applications.
var applicationsColumns = ApplicationsColumns{
	Id:          "id",
	Name:        "name",
	Alias:       "alias",
	ClientId:    "client_id",
	SecretKey:   "secret_key",
	AccessModel: "access_model",
	Status:      "status",
	CallbackUrl: "callback_url",
	Whitelist:   "whitelist",
	CreatedAt:   "created_at",
	UpdatedAt:   "updated_at",
	DeletedAt:   "deleted_at",
	CreateBy:    "create_by",
	UpdateBy:    "update_by",
}

// NewApplicationsDao creates and returns a new DAO object for table data access.
func NewApplicationsDao(handlers ...gdb.ModelHandler) *ApplicationsDao {
	return &ApplicationsDao{
		group:    "default",
		table:    "plugin_linapro_uidentity_cas_applications",
		columns:  applicationsColumns,
		handlers: handlers,
	}
}

// DB retrieves and returns the underlying raw database management object of the current DAO.
func (dao *ApplicationsDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of the current DAO.
func (dao *ApplicationsDao) Table() string {
	return dao.table
}

// Columns returns all column names of the current DAO.
func (dao *ApplicationsDao) Columns() ApplicationsColumns {
	return dao.columns
}

// Group returns the database configuration group name of the current DAO.
func (dao *ApplicationsDao) Group() string {
	return dao.group
}

// Ctx creates and returns a Model for the current DAO. It automatically sets the context for the current operation.
func (dao *ApplicationsDao) Ctx(ctx context.Context) *gdb.Model {
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
func (dao *ApplicationsDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
