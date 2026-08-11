// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// ActivationDao is the data access object for the table plugin_sicau_niu_activation.
type ActivationDao struct {
	table    string             // table is the underlying table name of the DAO.
	group    string             // group is the database configuration group name of the current DAO.
	columns  ActivationColumns  // columns contains all the column names of Table for convenient usage.
	handlers []gdb.ModelHandler // handlers for customized model modification.
}

// ActivationColumns defines and stores column names for the table plugin_sicau_niu_activation.
type ActivationColumns struct {
	Id           string //
	UserId       string // Activating player ID
	NiuId        string // Activated cattle ID
	ActivityDate string // Activation date YYYY-MM-DD for the per-day limit
	ActivatedAt  string // Activation time
	IsFirst      string // Whether this is the cattle first-activation: 1 first, 0 later
	OrderNo      string // Arrival order for this cattle, starting at 1
	PhotoPath    string // Optional activation photo storage path (evidence only, no recognition)
	CreatedAt    string // Creation time
	UpdatedAt    string // Update time
	DeletedAt    string // Soft-delete time, NULL means active
	RequestId    string // Player-scoped idempotency key
	ResponseJson string // Stable successful activation response for idempotent replay
}

// activationColumns holds the columns for the table plugin_sicau_niu_activation.
var activationColumns = ActivationColumns{
	Id:           "id",
	UserId:       "user_id",
	NiuId:        "niu_id",
	ActivityDate: "activity_date",
	ActivatedAt:  "activated_at",
	IsFirst:      "is_first",
	OrderNo:      "order_no",
	PhotoPath:    "photo_path",
	CreatedAt:    "created_at",
	UpdatedAt:    "updated_at",
	DeletedAt:    "deleted_at",
	RequestId:    "request_id",
	ResponseJson: "response_json",
}

// NewActivationDao creates and returns a new DAO object for table data access.
func NewActivationDao(handlers ...gdb.ModelHandler) *ActivationDao {
	return &ActivationDao{
		group:    "default",
		table:    "plugin_sicau_niu_activation",
		columns:  activationColumns,
		handlers: handlers,
	}
}

// DB retrieves and returns the underlying raw database management object of the current DAO.
func (dao *ActivationDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of the current DAO.
func (dao *ActivationDao) Table() string {
	return dao.table
}

// Columns returns all column names of the current DAO.
func (dao *ActivationDao) Columns() ActivationColumns {
	return dao.columns
}

// Group returns the database configuration group name of the current DAO.
func (dao *ActivationDao) Group() string {
	return dao.group
}

// Ctx creates and returns a Model for the current DAO. It automatically sets the context for the current operation.
func (dao *ActivationDao) Ctx(ctx context.Context) *gdb.Model {
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
func (dao *ActivationDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
