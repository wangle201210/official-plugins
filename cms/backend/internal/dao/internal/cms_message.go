// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// CmsMessageDao is the data access object for the table plugin_cms_message.
type CmsMessageDao struct {
	table    string             // table is the underlying table name of the DAO.
	group    string             // group is the database configuration group name of the current DAO.
	columns  CmsMessageColumns  // columns contains all the column names of Table for convenient usage.
	handlers []gdb.ModelHandler // handlers for customized model modification.
}

// CmsMessageColumns defines and stores column names for the table plugin_cms_message.
type CmsMessageColumns struct {
	Id        string // Message ID
	Name      string // Visitor name
	Mobile    string // Visitor mobile
	Email     string // Visitor email
	Content   string // Message content
	Reply     string // Reply content
	Status    string // Status: 0=pending, 1=approved, 2=rejected
	UserIp    string // Visitor IP
	UserAgent string // Visitor user agent
	CreatedBy string // Creator user ID
	UpdatedBy string // Updater user ID
	CreatedAt string // Creation time
	UpdatedAt string // Update time
	DeletedAt string // Deletion time
}

// cmsMessageColumns holds the columns for the table plugin_cms_message.
var cmsMessageColumns = CmsMessageColumns{
	Id:        "id",
	Name:      "name",
	Mobile:    "mobile",
	Email:     "email",
	Content:   "content",
	Reply:     "reply",
	Status:    "status",
	UserIp:    "user_ip",
	UserAgent: "user_agent",
	CreatedBy: "created_by",
	UpdatedBy: "updated_by",
	CreatedAt: "created_at",
	UpdatedAt: "updated_at",
	DeletedAt: "deleted_at",
}

// NewCmsMessageDao creates and returns a new DAO object for table data access.
func NewCmsMessageDao(handlers ...gdb.ModelHandler) *CmsMessageDao {
	return &CmsMessageDao{
		group:    "default",
		table:    "plugin_cms_message",
		columns:  cmsMessageColumns,
		handlers: handlers,
	}
}

// DB retrieves and returns the underlying raw database management object of the current DAO.
func (dao *CmsMessageDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of the current DAO.
func (dao *CmsMessageDao) Table() string {
	return dao.table
}

// Columns returns all column names of the current DAO.
func (dao *CmsMessageDao) Columns() CmsMessageColumns {
	return dao.columns
}

// Group returns the database configuration group name of the current DAO.
func (dao *CmsMessageDao) Group() string {
	return dao.group
}

// Ctx creates and returns a Model for the current DAO. It automatically sets the context for the current operation.
func (dao *CmsMessageDao) Ctx(ctx context.Context) *gdb.Model {
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
func (dao *CmsMessageDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
