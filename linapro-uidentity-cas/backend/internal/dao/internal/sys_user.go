// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// SysUserDao is the data access object for the table plugin_linapro_uidentity_cas_sys_user.
type SysUserDao struct {
	table    string             // table is the underlying table name of the DAO.
	group    string             // group is the database configuration group name of the current DAO.
	columns  SysUserColumns     // columns contains all the column names of Table for convenient usage.
	handlers []gdb.ModelHandler // handlers for customized model modification.
}

// SysUserColumns defines and stores column names for the table plugin_linapro_uidentity_cas_sys_user.
type SysUserColumns struct {
	UserId    string //
	Username  string //
	Password  string //
	NickName  string //
	Phone     string //
	RoleId    string //
	Salt      string //
	Avatar    string //
	Sex       string //
	Email     string //
	DeptId    string //
	PostId    string //
	Remark    string //
	Status    string //
	CreateBy  string //
	UpdateBy  string //
	CreatedAt string //
	UpdatedAt string //
	DeletedAt string //
}

// sysUserColumns holds the columns for the table plugin_linapro_uidentity_cas_sys_user.
var sysUserColumns = SysUserColumns{
	UserId:    "user_id",
	Username:  "username",
	Password:  "password",
	NickName:  "nick_name",
	Phone:     "phone",
	RoleId:    "role_id",
	Salt:      "salt",
	Avatar:    "avatar",
	Sex:       "sex",
	Email:     "email",
	DeptId:    "dept_id",
	PostId:    "post_id",
	Remark:    "remark",
	Status:    "status",
	CreateBy:  "create_by",
	UpdateBy:  "update_by",
	CreatedAt: "created_at",
	UpdatedAt: "updated_at",
	DeletedAt: "deleted_at",
}

// NewSysUserDao creates and returns a new DAO object for table data access.
func NewSysUserDao(handlers ...gdb.ModelHandler) *SysUserDao {
	return &SysUserDao{
		group:    "default",
		table:    "plugin_linapro_uidentity_cas_sys_user",
		columns:  sysUserColumns,
		handlers: handlers,
	}
}

// DB retrieves and returns the underlying raw database management object of the current DAO.
func (dao *SysUserDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of the current DAO.
func (dao *SysUserDao) Table() string {
	return dao.table
}

// Columns returns all column names of the current DAO.
func (dao *SysUserDao) Columns() SysUserColumns {
	return dao.columns
}

// Group returns the database configuration group name of the current DAO.
func (dao *SysUserDao) Group() string {
	return dao.group
}

// Ctx creates and returns a Model for the current DAO. It automatically sets the context for the current operation.
func (dao *SysUserDao) Ctx(ctx context.Context) *gdb.Model {
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
func (dao *SysUserDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
