// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// InboxMsgDao is the data access object for the table plugin_sicau_niu_inbox_msg.
type InboxMsgDao struct {
	table    string             // table is the underlying table name of the DAO.
	group    string             // group is the database configuration group name of the current DAO.
	columns  InboxMsgColumns    // columns contains all the column names of Table for convenient usage.
	handlers []gdb.ModelHandler // handlers for customized model modification.
}

// InboxMsgColumns defines and stores column names for the table plugin_sicau_niu_inbox_msg.
type InboxMsgColumns struct {
	Id        string //
	UserId    string //
	MsgType   string // Message type: stolen, gift_received
	Content   string //
	IsRead    string // Read flag: 1 read, 0 unread
	CreatedAt string //
	UpdatedAt string //
	DeletedAt string //
}

// inboxMsgColumns holds the columns for the table plugin_sicau_niu_inbox_msg.
var inboxMsgColumns = InboxMsgColumns{
	Id:        "id",
	UserId:    "user_id",
	MsgType:   "msg_type",
	Content:   "content",
	IsRead:    "is_read",
	CreatedAt: "created_at",
	UpdatedAt: "updated_at",
	DeletedAt: "deleted_at",
}

// NewInboxMsgDao creates and returns a new DAO object for table data access.
func NewInboxMsgDao(handlers ...gdb.ModelHandler) *InboxMsgDao {
	return &InboxMsgDao{
		group:    "default",
		table:    "plugin_sicau_niu_inbox_msg",
		columns:  inboxMsgColumns,
		handlers: handlers,
	}
}

// DB retrieves and returns the underlying raw database management object of the current DAO.
func (dao *InboxMsgDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of the current DAO.
func (dao *InboxMsgDao) Table() string {
	return dao.table
}

// Columns returns all column names of the current DAO.
func (dao *InboxMsgDao) Columns() InboxMsgColumns {
	return dao.columns
}

// Group returns the database configuration group name of the current DAO.
func (dao *InboxMsgDao) Group() string {
	return dao.group
}

// Ctx creates and returns a Model for the current DAO. It automatically sets the context for the current operation.
func (dao *InboxMsgDao) Ctx(ctx context.Context) *gdb.Model {
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
func (dao *InboxMsgDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
