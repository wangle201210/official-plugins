// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// CmsAlbumImageDao is the data access object for the table plugin_cms_album_image.
type CmsAlbumImageDao struct {
	table    string               // table is the underlying table name of the DAO.
	group    string               // group is the database configuration group name of the current DAO.
	columns  CmsAlbumImageColumns // columns contains all the column names of Table for convenient usage.
	handlers []gdb.ModelHandler   // handlers for customized model modification.
}

// CmsAlbumImageColumns defines and stores column names for the table plugin_cms_album_image.
type CmsAlbumImageColumns struct {
	Id        string // Image ID
	AlbumId   string // Album ID
	Url       string // Image URL
	Title     string // Image title
	Sort      string // Display order
	CreatedAt string // Creation time
	UpdatedAt string // Update time
}

// cmsAlbumImageColumns holds the columns for the table plugin_cms_album_image.
var cmsAlbumImageColumns = CmsAlbumImageColumns{
	Id:        "id",
	AlbumId:   "album_id",
	Url:       "url",
	Title:     "title",
	Sort:      "sort",
	CreatedAt: "created_at",
	UpdatedAt: "updated_at",
}

// NewCmsAlbumImageDao creates and returns a new DAO object for table data access.
func NewCmsAlbumImageDao(handlers ...gdb.ModelHandler) *CmsAlbumImageDao {
	return &CmsAlbumImageDao{
		group:    "default",
		table:    "plugin_cms_album_image",
		columns:  cmsAlbumImageColumns,
		handlers: handlers,
	}
}

// DB retrieves and returns the underlying raw database management object of the current DAO.
func (dao *CmsAlbumImageDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of the current DAO.
func (dao *CmsAlbumImageDao) Table() string {
	return dao.table
}

// Columns returns all column names of the current DAO.
func (dao *CmsAlbumImageDao) Columns() CmsAlbumImageColumns {
	return dao.columns
}

// Group returns the database configuration group name of the current DAO.
func (dao *CmsAlbumImageDao) Group() string {
	return dao.group
}

// Ctx creates and returns a Model for the current DAO. It automatically sets the context for the current operation.
func (dao *CmsAlbumImageDao) Ctx(ctx context.Context) *gdb.Model {
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
func (dao *CmsAlbumImageDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
