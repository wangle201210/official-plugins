// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// CmsSiteDao is the data access object for the table plugin_cms_site.
type CmsSiteDao struct {
	table    string             // table is the underlying table name of the DAO.
	group    string             // group is the database configuration group name of the current DAO.
	columns  CmsSiteColumns     // columns contains all the column names of Table for convenient usage.
	handlers []gdb.ModelHandler // handlers for customized model modification.
}

// CmsSiteColumns defines and stores column names for the table plugin_cms_site.
type CmsSiteColumns struct {
	Id           string // Site ID
	SiteKey      string // Stable site key
	Name         string // Site name
	Logo         string // Site logo URL
	Weixin       string // WeChat QR code image URL
	Domain       string // Primary site domain
	Slogan       string // Site slogan
	Keywords     string // SEO keywords
	Description  string // SEO description
	Icp          string // ICP record number
	Contact      string // Contact person
	Phone        string // Contact phone
	Email        string // Contact email
	Address      string // Contact address
	Status       string // Status: 0=disabled, 1=enabled
	CreatedBy    string // Creator user ID
	UpdatedBy    string // Updater user ID
	CreatedAt    string // Creation time
	UpdatedAt    string // Update time
	DeletedAt    string // Deletion time
	ShowMessages string // Show approved visitor messages on public message page: 0=no, 1=yes
}

// cmsSiteColumns holds the columns for the table plugin_cms_site.
var cmsSiteColumns = CmsSiteColumns{
	Id:           "id",
	SiteKey:      "site_key",
	Name:         "name",
	Logo:         "logo",
	Weixin:       "weixin",
	Domain:       "domain",
	Slogan:       "slogan",
	Keywords:     "keywords",
	Description:  "description",
	Icp:          "icp",
	Contact:      "contact",
	Phone:        "phone",
	Email:        "email",
	Address:      "address",
	Status:       "status",
	CreatedBy:    "created_by",
	UpdatedBy:    "updated_by",
	CreatedAt:    "created_at",
	UpdatedAt:    "updated_at",
	DeletedAt:    "deleted_at",
	ShowMessages: "show_messages",
}

// NewCmsSiteDao creates and returns a new DAO object for table data access.
func NewCmsSiteDao(handlers ...gdb.ModelHandler) *CmsSiteDao {
	return &CmsSiteDao{
		group:    "default",
		table:    "plugin_cms_site",
		columns:  cmsSiteColumns,
		handlers: handlers,
	}
}

// DB retrieves and returns the underlying raw database management object of the current DAO.
func (dao *CmsSiteDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of the current DAO.
func (dao *CmsSiteDao) Table() string {
	return dao.table
}

// Columns returns all column names of the current DAO.
func (dao *CmsSiteDao) Columns() CmsSiteColumns {
	return dao.columns
}

// Group returns the database configuration group name of the current DAO.
func (dao *CmsSiteDao) Group() string {
	return dao.group
}

// Ctx creates and returns a Model for the current DAO. It automatically sets the context for the current operation.
func (dao *CmsSiteDao) Ctx(ctx context.Context) *gdb.Model {
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
func (dao *CmsSiteDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
