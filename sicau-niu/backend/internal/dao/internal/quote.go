// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// QuoteDao is the data access object for the table plugin_sicau_niu_quote.
type QuoteDao struct {
	table    string             // table is the underlying table name of the DAO.
	group    string             // group is the database configuration group name of the current DAO.
	columns  QuoteColumns       // columns contains all the column names of Table for convenient usage.
	handlers []gdb.ModelHandler // handlers for customized model modification.
}

// QuoteColumns defines and stores column names for the table plugin_sicau_niu_quote.
type QuoteColumns struct {
	Id        string //
	Content   string // Quote text
	Enabled   string // Whether the quote participates in random playback: 1 enabled, 0 disabled
	CreatedAt string // Creation time
	UpdatedAt string // Update time
	DeletedAt string // Soft-delete time, NULL means active
}

// quoteColumns holds the columns for the table plugin_sicau_niu_quote.
var quoteColumns = QuoteColumns{
	Id:        "id",
	Content:   "content",
	Enabled:   "enabled",
	CreatedAt: "created_at",
	UpdatedAt: "updated_at",
	DeletedAt: "deleted_at",
}

// NewQuoteDao creates and returns a new DAO object for table data access.
func NewQuoteDao(handlers ...gdb.ModelHandler) *QuoteDao {
	return &QuoteDao{
		group:    "default",
		table:    "plugin_sicau_niu_quote",
		columns:  quoteColumns,
		handlers: handlers,
	}
}

// DB retrieves and returns the underlying raw database management object of the current DAO.
func (dao *QuoteDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of the current DAO.
func (dao *QuoteDao) Table() string {
	return dao.table
}

// Columns returns all column names of the current DAO.
func (dao *QuoteDao) Columns() QuoteColumns {
	return dao.columns
}

// Group returns the database configuration group name of the current DAO.
func (dao *QuoteDao) Group() string {
	return dao.group
}

// Ctx creates and returns a Model for the current DAO. It automatically sets the context for the current operation.
func (dao *QuoteDao) Ctx(ctx context.Context) *gdb.Model {
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
func (dao *QuoteDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
