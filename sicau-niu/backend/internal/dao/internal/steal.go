// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// StealDao is the data access object for the table plugin_sicau_niu_steal.
type StealDao struct {
	table    string             // table is the underlying table name of the DAO.
	group    string             // group is the database configuration group name of the current DAO.
	columns  StealColumns       // columns contains all the column names of Table for convenient usage.
	handlers []gdb.ModelHandler // handlers for customized model modification.
}

// StealColumns defines and stores column names for the table plugin_sicau_niu_steal.
type StealColumns struct {
	Id            string //
	ActorUserId   string //
	TargetUserId  string //
	Amount        string //
	StealDate     string // Steal date YYYY-MM-DD for the daily count limit
	CreatedAt     string //
	UpdatedAt     string //
	DeletedAt     string //
	RequestId     string // Required player-scoped idempotency key for stable replay
	ResultBalance string // Actor balance returned by the first successful steal request
}

// stealColumns holds the columns for the table plugin_sicau_niu_steal.
var stealColumns = StealColumns{
	Id:            "id",
	ActorUserId:   "actor_user_id",
	TargetUserId:  "target_user_id",
	Amount:        "amount",
	StealDate:     "steal_date",
	CreatedAt:     "created_at",
	UpdatedAt:     "updated_at",
	DeletedAt:     "deleted_at",
	RequestId:     "request_id",
	ResultBalance: "result_balance",
}

// NewStealDao creates and returns a new DAO object for table data access.
func NewStealDao(handlers ...gdb.ModelHandler) *StealDao {
	return &StealDao{
		group:    "default",
		table:    "plugin_sicau_niu_steal",
		columns:  stealColumns,
		handlers: handlers,
	}
}

// DB retrieves and returns the underlying raw database management object of the current DAO.
func (dao *StealDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of the current DAO.
func (dao *StealDao) Table() string {
	return dao.table
}

// Columns returns all column names of the current DAO.
func (dao *StealDao) Columns() StealColumns {
	return dao.columns
}

// Group returns the database configuration group name of the current DAO.
func (dao *StealDao) Group() string {
	return dao.group
}

// Ctx creates and returns a Model for the current DAO. It automatically sets the context for the current operation.
func (dao *StealDao) Ctx(ctx context.Context) *gdb.Model {
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
func (dao *StealDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
