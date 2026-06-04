// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// IronDao is the data access object for the table plugin_sicau_niu_iron.
type IronDao struct {
	table    string             // table is the underlying table name of the DAO.
	group    string             // group is the database configuration group name of the current DAO.
	columns  IronColumns        // columns contains all the column names of Table for convenient usage.
	handlers []gdb.ModelHandler // handlers for customized model modification.
}

// IronColumns defines and stores column names for the table plugin_sicau_niu_iron.
type IronColumns struct {
	Id        string //
	Code      string // Iron-cow device identifier, unique among active rows
	Name      string // Iron-cow display name
	LastLat   string // Latest GPS latitude pulled by the bonus flow (C4)
	LastLng   string // Latest GPS longitude pulled by the bonus flow (C4)
	LocatedAt string // Latest location pull time
	Remark    string // Iron-cow remark
	CreatedAt string // Creation time
	UpdatedAt string // Update time
	DeletedAt string // Soft-delete time, NULL means active
}

// ironColumns holds the columns for the table plugin_sicau_niu_iron.
var ironColumns = IronColumns{
	Id:        "id",
	Code:      "code",
	Name:      "name",
	LastLat:   "last_lat",
	LastLng:   "last_lng",
	LocatedAt: "located_at",
	Remark:    "remark",
	CreatedAt: "created_at",
	UpdatedAt: "updated_at",
	DeletedAt: "deleted_at",
}

// NewIronDao creates and returns a new DAO object for table data access.
func NewIronDao(handlers ...gdb.ModelHandler) *IronDao {
	return &IronDao{
		group:    "default",
		table:    "plugin_sicau_niu_iron",
		columns:  ironColumns,
		handlers: handlers,
	}
}

// DB retrieves and returns the underlying raw database management object of the current DAO.
func (dao *IronDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of the current DAO.
func (dao *IronDao) Table() string {
	return dao.table
}

// Columns returns all column names of the current DAO.
func (dao *IronDao) Columns() IronColumns {
	return dao.columns
}

// Group returns the database configuration group name of the current DAO.
func (dao *IronDao) Group() string {
	return dao.group
}

// Ctx creates and returns a Model for the current DAO. It automatically sets the context for the current operation.
func (dao *IronDao) Ctx(ctx context.Context) *gdb.Model {
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
func (dao *IronDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
