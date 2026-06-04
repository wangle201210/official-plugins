// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// CmsLinkDao is the data access object for the table plugin_cms_link.
type CmsLinkDao struct {
	table    string             // table is the underlying table name of the DAO.
	group    string             // group is the database configuration group name of the current DAO.
	columns  CmsLinkColumns     // columns contains all the column names of Table for convenient usage.
	handlers []gdb.ModelHandler // handlers for customized model modification.
}

// CmsLinkColumns defines and stores column names for the table plugin_cms_link.
type CmsLinkColumns struct {
	Id        string // Link ID
	GroupCode string // Display group code
	Name      string // Link name
	Url       string // Link URL
	Logo      string // Logo URL
	Sort      string // Display order
	Status    string // Status: 0=disabled, 1=enabled
	CreatedBy string // Creator user ID
	UpdatedBy string // Updater user ID
	CreatedAt string // Creation time
	UpdatedAt string // Update time
	DeletedAt string // Deletion time
}

// cmsLinkColumns holds the columns for the table plugin_cms_link.
var cmsLinkColumns = CmsLinkColumns{
	Id:        "id",
	GroupCode: "group_code",
	Name:      "name",
	Url:       "url",
	Logo:      "logo",
	Sort:      "sort",
	Status:    "status",
	CreatedBy: "created_by",
	UpdatedBy: "updated_by",
	CreatedAt: "created_at",
	UpdatedAt: "updated_at",
	DeletedAt: "deleted_at",
}

// NewCmsLinkDao creates and returns a new DAO object for table data access.
func NewCmsLinkDao(handlers ...gdb.ModelHandler) *CmsLinkDao {
	return &CmsLinkDao{
		group:    "default",
		table:    "plugin_cms_link",
		columns:  cmsLinkColumns,
		handlers: handlers,
	}
}

// DB retrieves and returns the underlying raw database management object of the current DAO.
func (dao *CmsLinkDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of the current DAO.
func (dao *CmsLinkDao) Table() string {
	return dao.table
}

// Columns returns all column names of the current DAO.
func (dao *CmsLinkDao) Columns() CmsLinkColumns {
	return dao.columns
}

// Group returns the database configuration group name of the current DAO.
func (dao *CmsLinkDao) Group() string {
	return dao.group
}

// Ctx creates and returns a Model for the current DAO. It automatically sets the context for the current operation.
func (dao *CmsLinkDao) Ctx(ctx context.Context) *gdb.Model {
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
func (dao *CmsLinkDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
