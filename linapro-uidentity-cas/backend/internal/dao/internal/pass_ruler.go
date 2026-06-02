// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// PassRulerDao is the data access object for the table plugin_linapro_uidentity_cas_pass_ruler.
type PassRulerDao struct {
	table    string             // table is the underlying table name of the DAO.
	group    string             // group is the database configuration group name of the current DAO.
	columns  PassRulerColumns   // columns contains all the column names of Table for convenient usage.
	handlers []gdb.ModelHandler // handlers for customized model modification.
}

// PassRulerColumns defines and stores column names for the table plugin_linapro_uidentity_cas_pass_ruler.
type PassRulerColumns struct {
	Id             string //
	Name           string //
	Capital        string //
	Lower          string //
	Number         string //
	Symbol         string //
	Length         string //
	Interval       string //
	IntervalStatus string //
	Status         string //
	CreatedAt      string //
	UpdatedAt      string //
	DeletedAt      string //
	CreateBy       string //
	UpdateBy       string //
}

// passRulerColumns holds the columns for the table plugin_linapro_uidentity_cas_pass_ruler.
var passRulerColumns = PassRulerColumns{
	Id:             "id",
	Name:           "name",
	Capital:        "capital",
	Lower:          "lower",
	Number:         "number",
	Symbol:         "symbol",
	Length:         "length",
	Interval:       "interval",
	IntervalStatus: "interval_status",
	Status:         "status",
	CreatedAt:      "created_at",
	UpdatedAt:      "updated_at",
	DeletedAt:      "deleted_at",
	CreateBy:       "create_by",
	UpdateBy:       "update_by",
}

// NewPassRulerDao creates and returns a new DAO object for table data access.
func NewPassRulerDao(handlers ...gdb.ModelHandler) *PassRulerDao {
	return &PassRulerDao{
		group:    "default",
		table:    "plugin_linapro_uidentity_cas_pass_ruler",
		columns:  passRulerColumns,
		handlers: handlers,
	}
}

// DB retrieves and returns the underlying raw database management object of the current DAO.
func (dao *PassRulerDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of the current DAO.
func (dao *PassRulerDao) Table() string {
	return dao.table
}

// Columns returns all column names of the current DAO.
func (dao *PassRulerDao) Columns() PassRulerColumns {
	return dao.columns
}

// Group returns the database configuration group name of the current DAO.
func (dao *PassRulerDao) Group() string {
	return dao.group
}

// Ctx creates and returns a Model for the current DAO. It automatically sets the context for the current operation.
func (dao *PassRulerDao) Ctx(ctx context.Context) *gdb.Model {
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
func (dao *PassRulerDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
