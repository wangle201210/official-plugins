// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// CmsProductDao is the data access object for the table plugin_cms_product.
type CmsProductDao struct {
	table    string             // table is the underlying table name of the DAO.
	group    string             // group is the database configuration group name of the current DAO.
	columns  CmsProductColumns  // columns contains all the column names of Table for convenient usage.
	handlers []gdb.ModelHandler // handlers for customized model modification.
}

// CmsProductColumns defines and stores column names for the table plugin_cms_product.
type CmsProductColumns struct {
	Id          string // Product ID
	CategoryId  string // Category ID
	Name        string // Product name
	Slug        string // Public URL slug
	Summary     string // Product summary
	Cover       string // Cover image URL
	Gallery     string // Gallery image URL list as JSON array text
	Price       string // Display price text
	Spec        string // Specification summary
	Content     string // Product detail HTML
	Keywords    string // SEO keywords
	Description string // SEO description
	Sort        string // Display order
	Status      string // Status: 0=draft, 1=published
	IsTop       string // Top flag: 0=no, 1=yes
	IsRecommend string // Recommend flag: 0=no, 1=yes
	Views       string // View count
	PublishedAt string // Publication time
	CreatedBy   string // Creator user ID
	UpdatedBy   string // Updater user ID
	CreatedAt   string // Creation time
	UpdatedAt   string // Update time
	DeletedAt   string // Deletion time
}

// cmsProductColumns holds the columns for the table plugin_cms_product.
var cmsProductColumns = CmsProductColumns{
	Id:          "id",
	CategoryId:  "category_id",
	Name:        "name",
	Slug:        "slug",
	Summary:     "summary",
	Cover:       "cover",
	Gallery:     "gallery",
	Price:       "price",
	Spec:        "spec",
	Content:     "content",
	Keywords:    "keywords",
	Description: "description",
	Sort:        "sort",
	Status:      "status",
	IsTop:       "is_top",
	IsRecommend: "is_recommend",
	Views:       "views",
	PublishedAt: "published_at",
	CreatedBy:   "created_by",
	UpdatedBy:   "updated_by",
	CreatedAt:   "created_at",
	UpdatedAt:   "updated_at",
	DeletedAt:   "deleted_at",
}

// NewCmsProductDao creates and returns a new DAO object for table data access.
func NewCmsProductDao(handlers ...gdb.ModelHandler) *CmsProductDao {
	return &CmsProductDao{
		group:    "default",
		table:    "plugin_cms_product",
		columns:  cmsProductColumns,
		handlers: handlers,
	}
}

// DB retrieves and returns the underlying raw database management object of the current DAO.
func (dao *CmsProductDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of the current DAO.
func (dao *CmsProductDao) Table() string {
	return dao.table
}

// Columns returns all column names of the current DAO.
func (dao *CmsProductDao) Columns() CmsProductColumns {
	return dao.columns
}

// Group returns the database configuration group name of the current DAO.
func (dao *CmsProductDao) Group() string {
	return dao.group
}

// Ctx creates and returns a Model for the current DAO. It automatically sets the context for the current operation.
func (dao *CmsProductDao) Ctx(ctx context.Context) *gdb.Model {
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
func (dao *CmsProductDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
