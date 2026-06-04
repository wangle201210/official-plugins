// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// GrassTxnDao is the data access object for the table plugin_sicau_niu_grass_txn.
type GrassTxnDao struct {
	table    string             // table is the underlying table name of the DAO.
	group    string             // group is the database configuration group name of the current DAO.
	columns  GrassTxnColumns    // columns contains all the column names of Table for convenient usage.
	handlers []gdb.ModelHandler // handlers for customized model modification.
}

// GrassTxnColumns defines and stores column names for the table plugin_sicau_niu_grass_txn.
type GrassTxnColumns struct {
	Id        string //
	UserId    string //
	Delta     string // Signed grass change: positive credit, negative debit
	TxnType   string // Type: checkin, feed, steal_gain, stolen_loss, gift_out, gift_in
	RefId     string // Related business record ID (feeding/steal/gift/checkin)
	CreatedAt string //
	UpdatedAt string //
	DeletedAt string //
}

// grassTxnColumns holds the columns for the table plugin_sicau_niu_grass_txn.
var grassTxnColumns = GrassTxnColumns{
	Id:        "id",
	UserId:    "user_id",
	Delta:     "delta",
	TxnType:   "txn_type",
	RefId:     "ref_id",
	CreatedAt: "created_at",
	UpdatedAt: "updated_at",
	DeletedAt: "deleted_at",
}

// NewGrassTxnDao creates and returns a new DAO object for table data access.
func NewGrassTxnDao(handlers ...gdb.ModelHandler) *GrassTxnDao {
	return &GrassTxnDao{
		group:    "default",
		table:    "plugin_sicau_niu_grass_txn",
		columns:  grassTxnColumns,
		handlers: handlers,
	}
}

// DB retrieves and returns the underlying raw database management object of the current DAO.
func (dao *GrassTxnDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of the current DAO.
func (dao *GrassTxnDao) Table() string {
	return dao.table
}

// Columns returns all column names of the current DAO.
func (dao *GrassTxnDao) Columns() GrassTxnColumns {
	return dao.columns
}

// Group returns the database configuration group name of the current DAO.
func (dao *GrassTxnDao) Group() string {
	return dao.group
}

// Ctx creates and returns a Model for the current DAO. It automatically sets the context for the current operation.
func (dao *GrassTxnDao) Ctx(ctx context.Context) *gdb.Model {
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
func (dao *GrassTxnDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
