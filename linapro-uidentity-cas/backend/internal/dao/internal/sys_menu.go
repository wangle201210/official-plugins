// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// SysMenuDao is the data access object for the table plugin_linapro_uidentity_cas_sys_menu.
type SysMenuDao struct {
	table    string             // table is the underlying table name of the DAO.
	group    string             // group is the database configuration group name of the current DAO.
	columns  SysMenuColumns     // columns contains all the column names of Table for convenient usage.
	handlers []gdb.ModelHandler // handlers for customized model modification.
}

// SysMenuColumns defines and stores column names for the table plugin_linapro_uidentity_cas_sys_menu.
type SysMenuColumns struct {
	MenuId     string //
	MenuName   string //
	Title      string //
	Icon       string //
	Path       string //
	Paths      string //
	MenuType   string //
	Action     string //
	Permission string //
	ParentId   string //
	NoCache    string //
	Breadcrumb string //
	Component  string //
	Sort       string //
	Visible    string //
	IsFrame    string //
	CreateBy   string //
	UpdateBy   string //
	CreatedAt  string //
	UpdatedAt  string //
	DeletedAt  string //
}

// sysMenuColumns holds the columns for the table plugin_linapro_uidentity_cas_sys_menu.
var sysMenuColumns = SysMenuColumns{
	MenuId:     "menu_id",
	MenuName:   "menu_name",
	Title:      "title",
	Icon:       "icon",
	Path:       "path",
	Paths:      "paths",
	MenuType:   "menu_type",
	Action:     "action",
	Permission: "permission",
	ParentId:   "parent_id",
	NoCache:    "no_cache",
	Breadcrumb: "breadcrumb",
	Component:  "component",
	Sort:       "sort",
	Visible:    "visible",
	IsFrame:    "is_frame",
	CreateBy:   "create_by",
	UpdateBy:   "update_by",
	CreatedAt:  "created_at",
	UpdatedAt:  "updated_at",
	DeletedAt:  "deleted_at",
}

// NewSysMenuDao creates and returns a new DAO object for table data access.
func NewSysMenuDao(handlers ...gdb.ModelHandler) *SysMenuDao {
	return &SysMenuDao{
		group:    "default",
		table:    "plugin_linapro_uidentity_cas_sys_menu",
		columns:  sysMenuColumns,
		handlers: handlers,
	}
}

// DB retrieves and returns the underlying raw database management object of the current DAO.
func (dao *SysMenuDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of the current DAO.
func (dao *SysMenuDao) Table() string {
	return dao.table
}

// Columns returns all column names of the current DAO.
func (dao *SysMenuDao) Columns() SysMenuColumns {
	return dao.columns
}

// Group returns the database configuration group name of the current DAO.
func (dao *SysMenuDao) Group() string {
	return dao.group
}

// Ctx creates and returns a Model for the current DAO. It automatically sets the context for the current operation.
func (dao *SysMenuDao) Ctx(ctx context.Context) *gdb.Model {
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
func (dao *SysMenuDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
