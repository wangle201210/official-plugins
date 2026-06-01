// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// JobLogDao is the data access object for the table plugin_linapro_uidentity_cas_job_log.
type JobLogDao struct {
	table    string             // table is the underlying table name of the DAO.
	group    string             // group is the database configuration group name of the current DAO.
	columns  JobLogColumns      // columns contains all the column names of Table for convenient usage.
	handlers []gdb.ModelHandler // handlers for customized model modification.
}

// JobLogColumns defines and stores column names for the table plugin_linapro_uidentity_cas_job_log.
type JobLogColumns struct {
	Id        string //
	TenantId  string //
	JobId     string //
	JobName   string //
	StartAt   string //
	EndAt     string //
	CreateNum string //
	UpdateNum string //
	DeleteNum string //
	ErrNum    string //
	CreatedBy string //
	UpdatedBy string //
	CreatedAt string //
	UpdatedAt string //
	DeletedAt string //
}

// jobLogColumns holds the columns for the table plugin_linapro_uidentity_cas_job_log.
var jobLogColumns = JobLogColumns{
	Id:        "id",
	TenantId:  "tenant_id",
	JobId:     "job_id",
	JobName:   "job_name",
	StartAt:   "start_at",
	EndAt:     "end_at",
	CreateNum: "create_num",
	UpdateNum: "update_num",
	DeleteNum: "delete_num",
	ErrNum:    "err_num",
	CreatedBy: "created_by",
	UpdatedBy: "updated_by",
	CreatedAt: "created_at",
	UpdatedAt: "updated_at",
	DeletedAt: "deleted_at",
}

// NewJobLogDao creates and returns a new DAO object for table data access.
func NewJobLogDao(handlers ...gdb.ModelHandler) *JobLogDao {
	return &JobLogDao{
		group:    "default",
		table:    "plugin_linapro_uidentity_cas_job_log",
		columns:  jobLogColumns,
		handlers: handlers,
	}
}

// DB retrieves and returns the underlying raw database management object of the current DAO.
func (dao *JobLogDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of the current DAO.
func (dao *JobLogDao) Table() string {
	return dao.table
}

// Columns returns all column names of the current DAO.
func (dao *JobLogDao) Columns() JobLogColumns {
	return dao.columns
}

// Group returns the database configuration group name of the current DAO.
func (dao *JobLogDao) Group() string {
	return dao.group
}

// Ctx creates and returns a Model for the current DAO. It automatically sets the context for the current operation.
func (dao *JobLogDao) Ctx(ctx context.Context) *gdb.Model {
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
func (dao *JobLogDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
