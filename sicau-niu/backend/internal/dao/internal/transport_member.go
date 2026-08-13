// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// TransportMemberDao is the data access object for the table plugin_sicau_niu_transport_member.
type TransportMemberDao struct {
	table    string                 // table is the underlying table name of the DAO.
	group    string                 // group is the database configuration group name of the current DAO.
	columns  TransportMemberColumns // columns contains all the column names of Table for convenient usage.
	handlers []gdb.ModelHandler     // handlers for customized model modification.
}

// TransportMemberColumns defines and stores column names for the table plugin_sicau_niu_transport_member.
type TransportMemberColumns struct {
	Id                      string //
	TeamId                  string //
	UserId                  string //
	JoinRequestId           string // Required joiner-scoped idempotency key; empty for creator membership
	Role                    string // Member display role: creator, member
	JoinedAt                string //
	LeftAt                  string //
	CreatedAt               string //
	UpdatedAt               string //
	TotalContributionMeters string //
	LastReportLat           string //
	LastReportLng           string //
	LastReportAt            string //
	JoinResponseJson        string // Stable first successful team join response for idempotent replay
}

// transportMemberColumns holds the columns for the table plugin_sicau_niu_transport_member.
var transportMemberColumns = TransportMemberColumns{
	Id:                      "id",
	TeamId:                  "team_id",
	UserId:                  "user_id",
	JoinRequestId:           "join_request_id",
	Role:                    "role",
	JoinedAt:                "joined_at",
	LeftAt:                  "left_at",
	CreatedAt:               "created_at",
	UpdatedAt:               "updated_at",
	TotalContributionMeters: "total_contribution_meters",
	LastReportLat:           "last_report_lat",
	LastReportLng:           "last_report_lng",
	LastReportAt:            "last_report_at",
	JoinResponseJson:        "join_response_json",
}

// NewTransportMemberDao creates and returns a new DAO object for table data access.
func NewTransportMemberDao(handlers ...gdb.ModelHandler) *TransportMemberDao {
	return &TransportMemberDao{
		group:    "default",
		table:    "plugin_sicau_niu_transport_member",
		columns:  transportMemberColumns,
		handlers: handlers,
	}
}

// DB retrieves and returns the underlying raw database management object of the current DAO.
func (dao *TransportMemberDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of the current DAO.
func (dao *TransportMemberDao) Table() string {
	return dao.table
}

// Columns returns all column names of the current DAO.
func (dao *TransportMemberDao) Columns() TransportMemberColumns {
	return dao.columns
}

// Group returns the database configuration group name of the current DAO.
func (dao *TransportMemberDao) Group() string {
	return dao.group
}

// Ctx creates and returns a Model for the current DAO. It automatically sets the context for the current operation.
func (dao *TransportMemberDao) Ctx(ctx context.Context) *gdb.Model {
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
func (dao *TransportMemberDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
