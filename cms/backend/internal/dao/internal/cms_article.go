// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// CmsArticleDao is the data access object for the table plugin_cms_article.
type CmsArticleDao struct {
	table    string             // table is the underlying table name of the DAO.
	group    string             // group is the database configuration group name of the current DAO.
	columns  CmsArticleColumns  // columns contains all the column names of Table for convenient usage.
	handlers []gdb.ModelHandler // handlers for customized model modification.
}

// CmsArticleColumns defines and stores column names for the table plugin_cms_article.
type CmsArticleColumns struct {
	Id          string // Article ID
	CategoryId  string // Category ID
	Title       string // Article title
	Subtitle    string // Article subtitle
	Slug        string // Public URL slug
	Summary     string // Article summary
	Cover       string // Cover image URL
	Author      string // Author name
	Source      string // Content source
	Content     string // Article body HTML
	Tags        string // Comma-separated tag names
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

// cmsArticleColumns holds the columns for the table plugin_cms_article.
var cmsArticleColumns = CmsArticleColumns{
	Id:          "id",
	CategoryId:  "category_id",
	Title:       "title",
	Subtitle:    "subtitle",
	Slug:        "slug",
	Summary:     "summary",
	Cover:       "cover",
	Author:      "author",
	Source:      "source",
	Content:     "content",
	Tags:        "tags",
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

// NewCmsArticleDao creates and returns a new DAO object for table data access.
func NewCmsArticleDao(handlers ...gdb.ModelHandler) *CmsArticleDao {
	return &CmsArticleDao{
		group:    "default",
		table:    "plugin_cms_article",
		columns:  cmsArticleColumns,
		handlers: handlers,
	}
}

// DB retrieves and returns the underlying raw database management object of the current DAO.
func (dao *CmsArticleDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of the current DAO.
func (dao *CmsArticleDao) Table() string {
	return dao.table
}

// Columns returns all column names of the current DAO.
func (dao *CmsArticleDao) Columns() CmsArticleColumns {
	return dao.columns
}

// Group returns the database configuration group name of the current DAO.
func (dao *CmsArticleDao) Group() string {
	return dao.group
}

// Ctx creates and returns a Model for the current DAO. It automatically sets the context for the current operation.
func (dao *CmsArticleDao) Ctx(ctx context.Context) *gdb.Model {
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
func (dao *CmsArticleDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
