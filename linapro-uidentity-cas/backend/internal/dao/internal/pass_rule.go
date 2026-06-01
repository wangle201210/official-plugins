// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// PassRuleDao is the data access object for the table plugin_linapro_uidentity_cas_pass_rule.
type PassRuleDao struct {
	table    string             // table is the underlying table name of the DAO.
	group    string             // group is the database configuration group name of the current DAO.
	columns  PassRuleColumns    // columns contains all the column names of Table for convenient usage.
	handlers []gdb.ModelHandler // handlers for customized model modification.
}

// PassRuleColumns defines and stores column names for the table plugin_linapro_uidentity_cas_pass_rule.
type PassRuleColumns struct {
	Id             string //
	TenantId       string //
	Name           string //
	Capital        string //
	Lower          string //
	Number         string //
	Symbol         string //
	Length         string //
	IntervalDays   string //
	IntervalStatus string //
	Status         string // Rule status: 0=disabled, 1=enabled
	CreatedBy      string //
	UpdatedBy      string //
	CreatedAt      string //
	UpdatedAt      string //
	DeletedAt      string //
}

// passRuleColumns holds the columns for the table plugin_linapro_uidentity_cas_pass_rule.
var passRuleColumns = PassRuleColumns{
	Id:             "id",
	TenantId:       "tenant_id",
	Name:           "name",
	Capital:        "capital",
	Lower:          "lower",
	Number:         "number",
	Symbol:         "symbol",
	Length:         "length",
	IntervalDays:   "interval_days",
	IntervalStatus: "interval_status",
	Status:         "status",
	CreatedBy:      "created_by",
	UpdatedBy:      "updated_by",
	CreatedAt:      "created_at",
	UpdatedAt:      "updated_at",
	DeletedAt:      "deleted_at",
}

// NewPassRuleDao creates and returns a new DAO object for table data access.
func NewPassRuleDao(handlers ...gdb.ModelHandler) *PassRuleDao {
	return &PassRuleDao{
		group:    "default",
		table:    "plugin_linapro_uidentity_cas_pass_rule",
		columns:  passRuleColumns,
		handlers: handlers,
	}
}

// DB retrieves and returns the underlying raw database management object of the current DAO.
func (dao *PassRuleDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of the current DAO.
func (dao *PassRuleDao) Table() string {
	return dao.table
}

// Columns returns all column names of the current DAO.
func (dao *PassRuleDao) Columns() PassRuleColumns {
	return dao.columns
}

// Group returns the database configuration group name of the current DAO.
func (dao *PassRuleDao) Group() string {
	return dao.group
}

// Ctx creates and returns a Model for the current DAO. It automatically sets the context for the current operation.
func (dao *PassRuleDao) Ctx(ctx context.Context) *gdb.Model {
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
func (dao *PassRuleDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
