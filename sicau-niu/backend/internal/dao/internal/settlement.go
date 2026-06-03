// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// SettlementDao is the data access object for the table plugin_sicau_niu_settlement.
type SettlementDao struct {
	table    string             // table is the underlying table name of the DAO.
	group    string             // group is the database configuration group name of the current DAO.
	columns  SettlementColumns  // columns contains all the column names of Table for convenient usage.
	handlers []gdb.ModelHandler // handlers for customized model modification.
}

// SettlementColumns defines and stores column names for the table plugin_sicau_niu_settlement.
type SettlementColumns struct {
	Id         string //
	Title      string // Archive title given by the operator
	Snapshot   string // Frozen dashboard metrics serialized as JSON text
	OperatorId string // Host operator user ID who created the archive
	ArchivedAt string // Archive time, set explicitly by the settlement service
	CreatedAt  string // Creation time
	UpdatedAt  string // Update time
	DeletedAt  string // Soft-delete time, NULL means active
}

// settlementColumns holds the columns for the table plugin_sicau_niu_settlement.
var settlementColumns = SettlementColumns{
	Id:         "id",
	Title:      "title",
	Snapshot:   "snapshot",
	OperatorId: "operator_id",
	ArchivedAt: "archived_at",
	CreatedAt:  "created_at",
	UpdatedAt:  "updated_at",
	DeletedAt:  "deleted_at",
}

// NewSettlementDao creates and returns a new DAO object for table data access.
func NewSettlementDao(handlers ...gdb.ModelHandler) *SettlementDao {
	return &SettlementDao{
		group:    "default",
		table:    "plugin_sicau_niu_settlement",
		columns:  settlementColumns,
		handlers: handlers,
	}
}

// DB retrieves and returns the underlying raw database management object of the current DAO.
func (dao *SettlementDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of the current DAO.
func (dao *SettlementDao) Table() string {
	return dao.table
}

// Columns returns all column names of the current DAO.
func (dao *SettlementDao) Columns() SettlementColumns {
	return dao.columns
}

// Group returns the database configuration group name of the current DAO.
func (dao *SettlementDao) Group() string {
	return dao.group
}

// Ctx creates and returns a Model for the current DAO. It automatically sets the context for the current operation.
func (dao *SettlementDao) Ctx(ctx context.Context) *gdb.Model {
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
func (dao *SettlementDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
