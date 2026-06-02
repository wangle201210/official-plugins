// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// SysConfigDao is the data access object for the table plugin_linapro_uidentity_cas_sys_config.
type SysConfigDao struct {
	table    string             // table is the underlying table name of the DAO.
	group    string             // group is the database configuration group name of the current DAO.
	columns  SysConfigColumns   // columns contains all the column names of Table for convenient usage.
	handlers []gdb.ModelHandler // handlers for customized model modification.
}

// SysConfigColumns defines and stores column names for the table plugin_linapro_uidentity_cas_sys_config.
type SysConfigColumns struct {
	Id          string //
	ConfigName  string //
	ConfigKey   string //
	ConfigValue string //
	ConfigType  string //
	IsFrontend  string //
	Remark      string //
	CreateBy    string //
	UpdateBy    string //
	CreatedAt   string //
	UpdatedAt   string //
	DeletedAt   string //
}

// sysConfigColumns holds the columns for the table plugin_linapro_uidentity_cas_sys_config.
var sysConfigColumns = SysConfigColumns{
	Id:          "id",
	ConfigName:  "config_name",
	ConfigKey:   "config_key",
	ConfigValue: "config_value",
	ConfigType:  "config_type",
	IsFrontend:  "is_frontend",
	Remark:      "remark",
	CreateBy:    "create_by",
	UpdateBy:    "update_by",
	CreatedAt:   "created_at",
	UpdatedAt:   "updated_at",
	DeletedAt:   "deleted_at",
}

// NewSysConfigDao creates and returns a new DAO object for table data access.
func NewSysConfigDao(handlers ...gdb.ModelHandler) *SysConfigDao {
	return &SysConfigDao{
		group:    "default",
		table:    "plugin_linapro_uidentity_cas_sys_config",
		columns:  sysConfigColumns,
		handlers: handlers,
	}
}

// DB retrieves and returns the underlying raw database management object of the current DAO.
func (dao *SysConfigDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of the current DAO.
func (dao *SysConfigDao) Table() string {
	return dao.table
}

// Columns returns all column names of the current DAO.
func (dao *SysConfigDao) Columns() SysConfigColumns {
	return dao.columns
}

// Group returns the database configuration group name of the current DAO.
func (dao *SysConfigDao) Group() string {
	return dao.group
}

// Ctx creates and returns a Model for the current DAO. It automatically sets the context for the current operation.
func (dao *SysConfigDao) Ctx(ctx context.Context) *gdb.Model {
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
func (dao *SysConfigDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
