// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// SysOperaLogDao is the data access object for the table plugin_linapro_uidentity_cas_sys_opera_log.
type SysOperaLogDao struct {
	table    string             // table is the underlying table name of the DAO.
	group    string             // group is the database configuration group name of the current DAO.
	columns  SysOperaLogColumns // columns contains all the column names of Table for convenient usage.
	handlers []gdb.ModelHandler // handlers for customized model modification.
}

// SysOperaLogColumns defines and stores column names for the table plugin_linapro_uidentity_cas_sys_opera_log.
type SysOperaLogColumns struct {
	Id            string //
	Title         string //
	BusinessType  string //
	BusinessTypes string //
	Method        string //
	RequestMethod string //
	OperatorType  string //
	OperName      string //
	DeptName      string //
	OperUrl       string //
	OperIp        string //
	OperLocation  string //
	OperParam     string //
	Status        string //
	OperTime      string //
	JsonResult    string //
	Remark        string //
	LatencyTime   string //
	UserAgent     string //
	CreatedAt     string //
	UpdatedAt     string //
	CreateBy      string //
	UpdateBy      string //
}

// sysOperaLogColumns holds the columns for the table plugin_linapro_uidentity_cas_sys_opera_log.
var sysOperaLogColumns = SysOperaLogColumns{
	Id:            "id",
	Title:         "title",
	BusinessType:  "business_type",
	BusinessTypes: "business_types",
	Method:        "method",
	RequestMethod: "request_method",
	OperatorType:  "operator_type",
	OperName:      "oper_name",
	DeptName:      "dept_name",
	OperUrl:       "oper_url",
	OperIp:        "oper_ip",
	OperLocation:  "oper_location",
	OperParam:     "oper_param",
	Status:        "status",
	OperTime:      "oper_time",
	JsonResult:    "json_result",
	Remark:        "remark",
	LatencyTime:   "latency_time",
	UserAgent:     "user_agent",
	CreatedAt:     "created_at",
	UpdatedAt:     "updated_at",
	CreateBy:      "create_by",
	UpdateBy:      "update_by",
}

// NewSysOperaLogDao creates and returns a new DAO object for table data access.
func NewSysOperaLogDao(handlers ...gdb.ModelHandler) *SysOperaLogDao {
	return &SysOperaLogDao{
		group:    "default",
		table:    "plugin_linapro_uidentity_cas_sys_opera_log",
		columns:  sysOperaLogColumns,
		handlers: handlers,
	}
}

// DB retrieves and returns the underlying raw database management object of the current DAO.
func (dao *SysOperaLogDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of the current DAO.
func (dao *SysOperaLogDao) Table() string {
	return dao.table
}

// Columns returns all column names of the current DAO.
func (dao *SysOperaLogDao) Columns() SysOperaLogColumns {
	return dao.columns
}

// Group returns the database configuration group name of the current DAO.
func (dao *SysOperaLogDao) Group() string {
	return dao.group
}

// Ctx creates and returns a Model for the current DAO. It automatically sets the context for the current operation.
func (dao *SysOperaLogDao) Ctx(ctx context.Context) *gdb.Model {
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
func (dao *SysOperaLogDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
