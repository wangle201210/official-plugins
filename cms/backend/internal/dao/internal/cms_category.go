// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// CmsCategoryDao is the data access object for the table plugin_cms_category.
type CmsCategoryDao struct {
	table    string             // table is the underlying table name of the DAO.
	group    string             // group is the database configuration group name of the current DAO.
	columns  CmsCategoryColumns // columns contains all the column names of Table for convenient usage.
	handlers []gdb.ModelHandler // handlers for customized model modification.
}

// CmsCategoryColumns defines and stores column names for the table plugin_cms_category.
type CmsCategoryColumns struct {
	Id              string // Category ID
	ParentId        string // Parent category ID
	Code            string // Stable category code
	Name            string // Category name
	Type            string // Category type: 1=list, 2=single page, 3=external link
	Path            string // Public category path
	ListTemplate    string // Public list template file
	ContentTemplate string // Public content/detail template file
	Cover           string // Category cover image URL
	Outlink         string // External link URL
	Title           string // SEO title
	Keywords        string // SEO keywords
	Description     string // SEO description
	Sort            string // Display order
	Status          string // Status: 0=disabled, 1=enabled
	CreatedBy       string // Creator user ID
	UpdatedBy       string // Updater user ID
	CreatedAt       string // Creation time
	UpdatedAt       string // Update time
	DeletedAt       string // Deletion time
}

// cmsCategoryColumns holds the columns for the table plugin_cms_category.
var cmsCategoryColumns = CmsCategoryColumns{
	Id:              "id",
	ParentId:        "parent_id",
	Code:            "code",
	Name:            "name",
	Type:            "type",
	Path:            "path",
	ListTemplate:    "list_template",
	ContentTemplate: "content_template",
	Cover:           "cover",
	Outlink:         "outlink",
	Title:           "title",
	Keywords:        "keywords",
	Description:     "description",
	Sort:            "sort",
	Status:          "status",
	CreatedBy:       "created_by",
	UpdatedBy:       "updated_by",
	CreatedAt:       "created_at",
	UpdatedAt:       "updated_at",
	DeletedAt:       "deleted_at",
}

// NewCmsCategoryDao creates and returns a new DAO object for table data access.
func NewCmsCategoryDao(handlers ...gdb.ModelHandler) *CmsCategoryDao {
	return &CmsCategoryDao{
		group:    "default",
		table:    "plugin_cms_category",
		columns:  cmsCategoryColumns,
		handlers: handlers,
	}
}

// DB retrieves and returns the underlying raw database management object of the current DAO.
func (dao *CmsCategoryDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of the current DAO.
func (dao *CmsCategoryDao) Table() string {
	return dao.table
}

// Columns returns all column names of the current DAO.
func (dao *CmsCategoryDao) Columns() CmsCategoryColumns {
	return dao.columns
}

// Group returns the database configuration group name of the current DAO.
func (dao *CmsCategoryDao) Group() string {
	return dao.group
}

// Ctx creates and returns a Model for the current DAO. It automatically sets the context for the current operation.
func (dao *CmsCategoryDao) Ctx(ctx context.Context) *gdb.Model {
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
func (dao *CmsCategoryDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
