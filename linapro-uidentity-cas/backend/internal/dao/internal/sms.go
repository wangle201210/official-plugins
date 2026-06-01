// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// SmsDao is the data access object for the table plugin_linapro_uidentity_cas_sms.
type SmsDao struct {
	table    string             // table is the underlying table name of the DAO.
	group    string             // group is the database configuration group name of the current DAO.
	columns  SmsColumns         // columns contains all the column names of Table for convenient usage.
	handlers []gdb.ModelHandler // handlers for customized model modification.
}

// SmsColumns defines and stores column names for the table plugin_linapro_uidentity_cas_sms.
type SmsColumns struct {
	Id        string //
	TenantId  string //
	Phone     string //
	Type      string //
	Content   string //
	Status    string //
	RespMsg   string //
	CreatedBy string //
	UpdatedBy string //
	CreatedAt string //
	UpdatedAt string //
	DeletedAt string //
}

// smsColumns holds the columns for the table plugin_linapro_uidentity_cas_sms.
var smsColumns = SmsColumns{
	Id:        "id",
	TenantId:  "tenant_id",
	Phone:     "phone",
	Type:      "type",
	Content:   "content",
	Status:    "status",
	RespMsg:   "resp_msg",
	CreatedBy: "created_by",
	UpdatedBy: "updated_by",
	CreatedAt: "created_at",
	UpdatedAt: "updated_at",
	DeletedAt: "deleted_at",
}

// NewSmsDao creates and returns a new DAO object for table data access.
func NewSmsDao(handlers ...gdb.ModelHandler) *SmsDao {
	return &SmsDao{
		group:    "default",
		table:    "plugin_linapro_uidentity_cas_sms",
		columns:  smsColumns,
		handlers: handlers,
	}
}

// DB retrieves and returns the underlying raw database management object of the current DAO.
func (dao *SmsDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of the current DAO.
func (dao *SmsDao) Table() string {
	return dao.table
}

// Columns returns all column names of the current DAO.
func (dao *SmsDao) Columns() SmsColumns {
	return dao.columns
}

// Group returns the database configuration group name of the current DAO.
func (dao *SmsDao) Group() string {
	return dao.group
}

// Ctx creates and returns a Model for the current DAO. It automatically sets the context for the current operation.
func (dao *SmsDao) Ctx(ctx context.Context) *gdb.Model {
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
func (dao *SmsDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
