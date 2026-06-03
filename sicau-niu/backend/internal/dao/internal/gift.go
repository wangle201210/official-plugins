// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// GiftDao is the data access object for the table plugin_sicau_niu_gift.
type GiftDao struct {
	table    string             // table is the underlying table name of the DAO.
	group    string             // group is the database configuration group name of the current DAO.
	columns  GiftColumns        // columns contains all the column names of Table for convenient usage.
	handlers []gdb.ModelHandler // handlers for customized model modification.
}

// GiftColumns defines and stores column names for the table plugin_sicau_niu_gift.
type GiftColumns struct {
	Id         string //
	FromUserId string //
	ToUserId   string //
	Amount     string //
	GiftDate   string // Gift date YYYY-MM-DD for the daily count limit
	CreatedAt  string //
	UpdatedAt  string //
	DeletedAt  string //
}

// giftColumns holds the columns for the table plugin_sicau_niu_gift.
var giftColumns = GiftColumns{
	Id:         "id",
	FromUserId: "from_user_id",
	ToUserId:   "to_user_id",
	Amount:     "amount",
	GiftDate:   "gift_date",
	CreatedAt:  "created_at",
	UpdatedAt:  "updated_at",
	DeletedAt:  "deleted_at",
}

// NewGiftDao creates and returns a new DAO object for table data access.
func NewGiftDao(handlers ...gdb.ModelHandler) *GiftDao {
	return &GiftDao{
		group:    "default",
		table:    "plugin_sicau_niu_gift",
		columns:  giftColumns,
		handlers: handlers,
	}
}

// DB retrieves and returns the underlying raw database management object of the current DAO.
func (dao *GiftDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of the current DAO.
func (dao *GiftDao) Table() string {
	return dao.table
}

// Columns returns all column names of the current DAO.
func (dao *GiftDao) Columns() GiftColumns {
	return dao.columns
}

// Group returns the database configuration group name of the current DAO.
func (dao *GiftDao) Group() string {
	return dao.group
}

// Ctx creates and returns a Model for the current DAO. It automatically sets the context for the current operation.
func (dao *GiftDao) Ctx(ctx context.Context) *gdb.Model {
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
func (dao *GiftDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
