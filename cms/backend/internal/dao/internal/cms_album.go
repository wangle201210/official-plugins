// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// CmsAlbumDao is the data access object for the table plugin_cms_album.
type CmsAlbumDao struct {
	table    string             // table is the underlying table name of the DAO.
	group    string             // group is the database configuration group name of the current DAO.
	columns  CmsAlbumColumns    // columns contains all the column names of Table for convenient usage.
	handlers []gdb.ModelHandler // handlers for customized model modification.
}

// CmsAlbumColumns defines and stores column names for the table plugin_cms_album.
type CmsAlbumColumns struct {
	Id          string // Album ID
	CategoryId  string // Category ID
	Name        string // Album name
	Cover       string // Cover image URL
	Description string // Album description
	Sort        string // Display order
	Status      string // Status: 0=disabled, 1=enabled
	CreatedBy   string // Creator user ID
	UpdatedBy   string // Updater user ID
	CreatedAt   string // Creation time
	UpdatedAt   string // Update time
	DeletedAt   string // Deletion time
}

// cmsAlbumColumns holds the columns for the table plugin_cms_album.
var cmsAlbumColumns = CmsAlbumColumns{
	Id:          "id",
	CategoryId:  "category_id",
	Name:        "name",
	Cover:       "cover",
	Description: "description",
	Sort:        "sort",
	Status:      "status",
	CreatedBy:   "created_by",
	UpdatedBy:   "updated_by",
	CreatedAt:   "created_at",
	UpdatedAt:   "updated_at",
	DeletedAt:   "deleted_at",
}

// NewCmsAlbumDao creates and returns a new DAO object for table data access.
func NewCmsAlbumDao(handlers ...gdb.ModelHandler) *CmsAlbumDao {
	return &CmsAlbumDao{
		group:    "default",
		table:    "plugin_cms_album",
		columns:  cmsAlbumColumns,
		handlers: handlers,
	}
}

// DB retrieves and returns the underlying raw database management object of the current DAO.
func (dao *CmsAlbumDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of the current DAO.
func (dao *CmsAlbumDao) Table() string {
	return dao.table
}

// Columns returns all column names of the current DAO.
func (dao *CmsAlbumDao) Columns() CmsAlbumColumns {
	return dao.columns
}

// Group returns the database configuration group name of the current DAO.
func (dao *CmsAlbumDao) Group() string {
	return dao.group
}

// Ctx creates and returns a Model for the current DAO. It automatically sets the context for the current operation.
func (dao *CmsAlbumDao) Ctx(ctx context.Context) *gdb.Model {
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
func (dao *CmsAlbumDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
