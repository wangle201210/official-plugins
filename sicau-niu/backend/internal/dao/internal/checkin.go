// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// CheckinDao is the data access object for the table plugin_sicau_niu_checkin.
type CheckinDao struct {
	table    string             // table is the underlying table name of the DAO.
	group    string             // group is the database configuration group name of the current DAO.
	columns  CheckinColumns     // columns contains all the column names of Table for convenient usage.
	handlers []gdb.ModelHandler // handlers for customized model modification.
}

// CheckinColumns defines and stores column names for the table plugin_sicau_niu_checkin.
type CheckinColumns struct {
	Id            string //
	UserId        string //
	CheckinDate   string // Check-in date YYYY-MM-DD
	Amount        string // Granted grass amount (20-50)
	CreatedAt     string //
	UpdatedAt     string //
	DeletedAt     string //
	RequestId     string // Player-scoped idempotency key
	ResultBalance string // Stable account balance returned by the first successful request
}

// checkinColumns holds the columns for the table plugin_sicau_niu_checkin.
var checkinColumns = CheckinColumns{
	Id:            "id",
	UserId:        "user_id",
	CheckinDate:   "checkin_date",
	Amount:        "amount",
	CreatedAt:     "created_at",
	UpdatedAt:     "updated_at",
	DeletedAt:     "deleted_at",
	RequestId:     "request_id",
	ResultBalance: "result_balance",
}

// NewCheckinDao creates and returns a new DAO object for table data access.
func NewCheckinDao(handlers ...gdb.ModelHandler) *CheckinDao {
	return &CheckinDao{
		group:    "default",
		table:    "plugin_sicau_niu_checkin",
		columns:  checkinColumns,
		handlers: handlers,
	}
}

// DB retrieves and returns the underlying raw database management object of the current DAO.
func (dao *CheckinDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of the current DAO.
func (dao *CheckinDao) Table() string {
	return dao.table
}

// Columns returns all column names of the current DAO.
func (dao *CheckinDao) Columns() CheckinColumns {
	return dao.columns
}

// Group returns the database configuration group name of the current DAO.
func (dao *CheckinDao) Group() string {
	return dao.group
}

// Ctx creates and returns a Model for the current DAO. It automatically sets the context for the current operation.
func (dao *CheckinDao) Ctx(ctx context.Context) *gdb.Model {
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
func (dao *CheckinDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
