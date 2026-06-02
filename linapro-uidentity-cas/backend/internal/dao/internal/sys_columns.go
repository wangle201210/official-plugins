// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// SysColumnsDao is the data access object for the table plugin_linapro_uidentity_cas_sys_columns.
type SysColumnsDao struct {
	table    string             // table is the underlying table name of the DAO.
	group    string             // group is the database configuration group name of the current DAO.
	columns  SysColumnsColumns  // columns contains all the column names of Table for convenient usage.
	handlers []gdb.ModelHandler // handlers for customized model modification.
}

// SysColumnsColumns defines and stores column names for the table plugin_linapro_uidentity_cas_sys_columns.
type SysColumnsColumns struct {
	ColumnId           string //
	TableId            string //
	ColumnName         string //
	ColumnComment      string //
	ColumnType         string //
	GoType             string //
	GoField            string //
	JsonField          string //
	IsPk               string //
	IsIncrement        string //
	IsRequired         string //
	IsInsert           string //
	IsEdit             string //
	IsList             string //
	IsQuery            string //
	QueryType          string //
	HtmlType           string //
	DictType           string //
	Sort               string //
	List               string //
	Pk                 string //
	Required           string //
	SuperColumn        string //
	UsableColumn       string //
	Increment          string //
	Insert             string //
	Edit               string //
	Query              string //
	Remark             string //
	FkTableName        string //
	FkTableNameClass   string //
	FkTableNamePackage string //
	FkLabelId          string //
	FkLabelName        string //
	CreateBy           string //
	UpdateBy           string //
	CreatedAt          string //
	UpdatedAt          string //
	DeletedAt          string //
}

// sysColumnsColumns holds the columns for the table plugin_linapro_uidentity_cas_sys_columns.
var sysColumnsColumns = SysColumnsColumns{
	ColumnId:           "column_id",
	TableId:            "table_id",
	ColumnName:         "column_name",
	ColumnComment:      "column_comment",
	ColumnType:         "column_type",
	GoType:             "go_type",
	GoField:            "go_field",
	JsonField:          "json_field",
	IsPk:               "is_pk",
	IsIncrement:        "is_increment",
	IsRequired:         "is_required",
	IsInsert:           "is_insert",
	IsEdit:             "is_edit",
	IsList:             "is_list",
	IsQuery:            "is_query",
	QueryType:          "query_type",
	HtmlType:           "html_type",
	DictType:           "dict_type",
	Sort:               "sort",
	List:               "list",
	Pk:                 "pk",
	Required:           "required",
	SuperColumn:        "super_column",
	UsableColumn:       "usable_column",
	Increment:          "increment",
	Insert:             "insert",
	Edit:               "edit",
	Query:              "query",
	Remark:             "remark",
	FkTableName:        "fk_table_name",
	FkTableNameClass:   "fk_table_name_class",
	FkTableNamePackage: "fk_table_name_package",
	FkLabelId:          "fk_label_id",
	FkLabelName:        "fk_label_name",
	CreateBy:           "create_by",
	UpdateBy:           "update_by",
	CreatedAt:          "created_at",
	UpdatedAt:          "updated_at",
	DeletedAt:          "deleted_at",
}

// NewSysColumnsDao creates and returns a new DAO object for table data access.
func NewSysColumnsDao(handlers ...gdb.ModelHandler) *SysColumnsDao {
	return &SysColumnsDao{
		group:    "default",
		table:    "plugin_linapro_uidentity_cas_sys_columns",
		columns:  sysColumnsColumns,
		handlers: handlers,
	}
}

// DB retrieves and returns the underlying raw database management object of the current DAO.
func (dao *SysColumnsDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of the current DAO.
func (dao *SysColumnsDao) Table() string {
	return dao.table
}

// Columns returns all column names of the current DAO.
func (dao *SysColumnsDao) Columns() SysColumnsColumns {
	return dao.columns
}

// Group returns the database configuration group name of the current DAO.
func (dao *SysColumnsDao) Group() string {
	return dao.group
}

// Ctx creates and returns a Model for the current DAO. It automatically sets the context for the current operation.
func (dao *SysColumnsDao) Ctx(ctx context.Context) *gdb.Model {
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
func (dao *SysColumnsDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
