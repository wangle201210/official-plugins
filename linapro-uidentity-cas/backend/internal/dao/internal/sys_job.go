// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// SysJobDao is the data access object for the table plugin_linapro_uidentity_cas_sys_job.
type SysJobDao struct {
	table    string             // table is the underlying table name of the DAO.
	group    string             // group is the database configuration group name of the current DAO.
	columns  SysJobColumns      // columns contains all the column names of Table for convenient usage.
	handlers []gdb.ModelHandler // handlers for customized model modification.
}

// SysJobColumns defines and stores column names for the table plugin_linapro_uidentity_cas_sys_job.
type SysJobColumns struct {
	JobId          string //
	TenantId       string //
	JobName        string //
	JobGroup       string //
	JobType        string // Job type: 1=http, 2=exec or plugin-defined executor
	CronExpression string //
	InvokeTarget   string //
	Args           string //
	MisfirePolicy  string // Misfire policy copied from legacy sys_job semantics
	Concurrent     string // Concurrent execution flag: 0=disallow, 1=allow
	Status         string // Job status: 1=disabled, 2=enabled
	EntryId        string // Runtime scheduler entry ID, 0 means not scheduled
	CreatedBy      string //
	UpdatedBy      string //
	CreatedAt      string //
	UpdatedAt      string //
	DeletedAt      string //
}

// sysJobColumns holds the columns for the table plugin_linapro_uidentity_cas_sys_job.
var sysJobColumns = SysJobColumns{
	JobId:          "job_id",
	TenantId:       "tenant_id",
	JobName:        "job_name",
	JobGroup:       "job_group",
	JobType:        "job_type",
	CronExpression: "cron_expression",
	InvokeTarget:   "invoke_target",
	Args:           "args",
	MisfirePolicy:  "misfire_policy",
	Concurrent:     "concurrent",
	Status:         "status",
	EntryId:        "entry_id",
	CreatedBy:      "created_by",
	UpdatedBy:      "updated_by",
	CreatedAt:      "created_at",
	UpdatedAt:      "updated_at",
	DeletedAt:      "deleted_at",
}

// NewSysJobDao creates and returns a new DAO object for table data access.
func NewSysJobDao(handlers ...gdb.ModelHandler) *SysJobDao {
	return &SysJobDao{
		group:    "default",
		table:    "plugin_linapro_uidentity_cas_sys_job",
		columns:  sysJobColumns,
		handlers: handlers,
	}
}

// DB retrieves and returns the underlying raw database management object of the current DAO.
func (dao *SysJobDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of the current DAO.
func (dao *SysJobDao) Table() string {
	return dao.table
}

// Columns returns all column names of the current DAO.
func (dao *SysJobDao) Columns() SysJobColumns {
	return dao.columns
}

// Group returns the database configuration group name of the current DAO.
func (dao *SysJobDao) Group() string {
	return dao.group
}

// Ctx creates and returns a Model for the current DAO. It automatically sets the context for the current operation.
func (dao *SysJobDao) Ctx(ctx context.Context) *gdb.Model {
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
func (dao *SysJobDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
