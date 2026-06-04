// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// CmsSlideDao is the data access object for the table plugin_cms_slide.
type CmsSlideDao struct {
	table    string             // table is the underlying table name of the DAO.
	group    string             // group is the database configuration group name of the current DAO.
	columns  CmsSlideColumns    // columns contains all the column names of Table for convenient usage.
	handlers []gdb.ModelHandler // handlers for customized model modification.
}

// CmsSlideColumns defines and stores column names for the table plugin_cms_slide.
type CmsSlideColumns struct {
	Id        string // Slide ID
	GroupCode string // Display group code
	Title     string // Slide title
	Subtitle  string // Slide subtitle
	Image     string // Slide image URL
	Link      string // Click target URL
	Sort      string // Display order
	Status    string // Status: 0=disabled, 1=enabled
	CreatedBy string // Creator user ID
	UpdatedBy string // Updater user ID
	CreatedAt string // Creation time
	UpdatedAt string // Update time
	DeletedAt string // Deletion time
}

// cmsSlideColumns holds the columns for the table plugin_cms_slide.
var cmsSlideColumns = CmsSlideColumns{
	Id:        "id",
	GroupCode: "group_code",
	Title:     "title",
	Subtitle:  "subtitle",
	Image:     "image",
	Link:      "link",
	Sort:      "sort",
	Status:    "status",
	CreatedBy: "created_by",
	UpdatedBy: "updated_by",
	CreatedAt: "created_at",
	UpdatedAt: "updated_at",
	DeletedAt: "deleted_at",
}

// NewCmsSlideDao creates and returns a new DAO object for table data access.
func NewCmsSlideDao(handlers ...gdb.ModelHandler) *CmsSlideDao {
	return &CmsSlideDao{
		group:    "default",
		table:    "plugin_cms_slide",
		columns:  cmsSlideColumns,
		handlers: handlers,
	}
}

// DB retrieves and returns the underlying raw database management object of the current DAO.
func (dao *CmsSlideDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of the current DAO.
func (dao *CmsSlideDao) Table() string {
	return dao.table
}

// Columns returns all column names of the current DAO.
func (dao *CmsSlideDao) Columns() CmsSlideColumns {
	return dao.columns
}

// Group returns the database configuration group name of the current DAO.
func (dao *CmsSlideDao) Group() string {
	return dao.group
}

// Ctx creates and returns a Model for the current DAO. It automatically sets the context for the current operation.
func (dao *CmsSlideDao) Ctx(ctx context.Context) *gdb.Model {
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
func (dao *CmsSlideDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
