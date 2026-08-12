// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// TransportReportDao is the data access object for the table plugin_sicau_niu_transport_report.
type TransportReportDao struct {
	table    string                 // table is the underlying table name of the DAO.
	group    string                 // group is the database configuration group name of the current DAO.
	columns  TransportReportColumns // columns contains all the column names of Table for convenient usage.
	handlers []gdb.ModelHandler     // handlers for customized model modification.
}

// TransportReportColumns defines and stores column names for the table plugin_sicau_niu_transport_report.
type TransportReportColumns struct {
	Id                 string //
	ActivityKey        string //
	TeamId             string //
	MemberId           string //
	UserId             string //
	RequestId          string //
	ActivityDate       string // Beijing natural-day key derived from accepted_at
	StartLat           string // Previous successful report latitude; NULL for the first report in a membership
	StartLng           string //
	EndLat             string //
	EndLng             string //
	SampledAt          string //
	AcceptedAt         string //
	ContributionMeters string //
	UserTotalMeters    string //
	TeamTotalMeters    string //
	DailyReportCount   string //
	CreatedAt          string //
}

// transportReportColumns holds the columns for the table plugin_sicau_niu_transport_report.
var transportReportColumns = TransportReportColumns{
	Id:                 "id",
	ActivityKey:        "activity_key",
	TeamId:             "team_id",
	MemberId:           "member_id",
	UserId:             "user_id",
	RequestId:          "request_id",
	ActivityDate:       "activity_date",
	StartLat:           "start_lat",
	StartLng:           "start_lng",
	EndLat:             "end_lat",
	EndLng:             "end_lng",
	SampledAt:          "sampled_at",
	AcceptedAt:         "accepted_at",
	ContributionMeters: "contribution_meters",
	UserTotalMeters:    "user_total_meters",
	TeamTotalMeters:    "team_total_meters",
	DailyReportCount:   "daily_report_count",
	CreatedAt:          "created_at",
}

// NewTransportReportDao creates and returns a new DAO object for table data access.
func NewTransportReportDao(handlers ...gdb.ModelHandler) *TransportReportDao {
	return &TransportReportDao{
		group:    "default",
		table:    "plugin_sicau_niu_transport_report",
		columns:  transportReportColumns,
		handlers: handlers,
	}
}

// DB retrieves and returns the underlying raw database management object of the current DAO.
func (dao *TransportReportDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of the current DAO.
func (dao *TransportReportDao) Table() string {
	return dao.table
}

// Columns returns all column names of the current DAO.
func (dao *TransportReportDao) Columns() TransportReportColumns {
	return dao.columns
}

// Group returns the database configuration group name of the current DAO.
func (dao *TransportReportDao) Group() string {
	return dao.group
}

// Ctx creates and returns a Model for the current DAO. It automatically sets the context for the current operation.
func (dao *TransportReportDao) Ctx(ctx context.Context) *gdb.Model {
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
func (dao *TransportReportDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
