// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// SysTablesDao is the data access object for the table plugin_linapro_uidentity_cas_sys_tables.
type SysTablesDao struct {
	table    string             // table is the underlying table name of the DAO.
	group    string             // group is the database configuration group name of the current DAO.
	columns  SysTablesColumns   // columns contains all the column names of Table for convenient usage.
	handlers []gdb.ModelHandler // handlers for customized model modification.
}

// SysTablesColumns defines and stores column names for the table plugin_linapro_uidentity_cas_sys_tables.
type SysTablesColumns struct {
	TableId             string //
	TableName           string //
	TableComment        string //
	ClassName           string //
	TplCategory         string //
	PackageName         string //
	ModuleName          string //
	ModuleFrontName     string //
	BusinessName        string //
	FunctionName        string //
	FunctionAuthor      string //
	PkColumn            string //
	PkGoField           string //
	PkJsonField         string //
	Options             string //
	TreeCode            string //
	TreeParentCode      string //
	TreeName            string //
	Tree                string //
	Crud                string //
	Remark              string //
	IsDataScope         string //
	IsActions           string //
	IsAuth              string //
	IsLogicalDelete     string //
	LogicalDelete       string //
	LogicalDeleteColumn string //
	CreateBy            string //
	UpdateBy            string //
	CreatedAt           string //
	UpdatedAt           string //
	DeletedAt           string //
}

// sysTablesColumns holds the columns for the table plugin_linapro_uidentity_cas_sys_tables.
var sysTablesColumns = SysTablesColumns{
	TableId:             "table_id",
	TableName:           "table_name",
	TableComment:        "table_comment",
	ClassName:           "class_name",
	TplCategory:         "tpl_category",
	PackageName:         "package_name",
	ModuleName:          "module_name",
	ModuleFrontName:     "module_front_name",
	BusinessName:        "business_name",
	FunctionName:        "function_name",
	FunctionAuthor:      "function_author",
	PkColumn:            "pk_column",
	PkGoField:           "pk_go_field",
	PkJsonField:         "pk_json_field",
	Options:             "options",
	TreeCode:            "tree_code",
	TreeParentCode:      "tree_parent_code",
	TreeName:            "tree_name",
	Tree:                "tree",
	Crud:                "crud",
	Remark:              "remark",
	IsDataScope:         "is_data_scope",
	IsActions:           "is_actions",
	IsAuth:              "is_auth",
	IsLogicalDelete:     "is_logical_delete",
	LogicalDelete:       "logical_delete",
	LogicalDeleteColumn: "logical_delete_column",
	CreateBy:            "create_by",
	UpdateBy:            "update_by",
	CreatedAt:           "created_at",
	UpdatedAt:           "updated_at",
	DeletedAt:           "deleted_at",
}

// NewSysTablesDao creates and returns a new DAO object for table data access.
func NewSysTablesDao(handlers ...gdb.ModelHandler) *SysTablesDao {
	return &SysTablesDao{
		group:    "default",
		table:    "plugin_linapro_uidentity_cas_sys_tables",
		columns:  sysTablesColumns,
		handlers: handlers,
	}
}

// DB retrieves and returns the underlying raw database management object of the current DAO.
func (dao *SysTablesDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of the current DAO.
func (dao *SysTablesDao) Table() string {
	return dao.table
}

// Columns returns all column names of the current DAO.
func (dao *SysTablesDao) Columns() SysTablesColumns {
	return dao.columns
}

// Group returns the database configuration group name of the current DAO.
func (dao *SysTablesDao) Group() string {
	return dao.group
}

// Ctx creates and returns a Model for the current DAO. It automatically sets the context for the current operation.
func (dao *SysTablesDao) Ctx(ctx context.Context) *gdb.Model {
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
func (dao *SysTablesDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
