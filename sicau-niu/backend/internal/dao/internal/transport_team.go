// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// TransportTeamDao is the data access object for the table plugin_sicau_niu_transport_team.
type TransportTeamDao struct {
	table    string               // table is the underlying table name of the DAO.
	group    string               // group is the database configuration group name of the current DAO.
	columns  TransportTeamColumns // columns contains all the column names of Table for convenient usage.
	handlers []gdb.ModelHandler   // handlers for customized model modification.
}

// TransportTeamColumns defines and stores column names for the table plugin_sicau_niu_transport_team.
type TransportTeamColumns struct {
	Id                      string //
	Name                    string //
	LeaderUserId            string // Team creator player ID; creator has no lifecycle authority
	CreateRequestId         string // Required creator-scoped idempotency key for stable replay
	Status                  string // Team status: effective, invalid
	Visible                 string //
	CreatedAt               string //
	UpdatedAt               string //
	DeletedAt               string //
	MemberCount             string //
	TotalContributionMeters string //
	LastActiveAt            string // Latest successful create, join or contribution server commit time
	InvalidatedAt           string //
	InvalidReason           string //
	CreateResponseJson      string // Stable first successful team creation response for idempotent replay
}

// transportTeamColumns holds the columns for the table plugin_sicau_niu_transport_team.
var transportTeamColumns = TransportTeamColumns{
	Id:                      "id",
	Name:                    "name",
	LeaderUserId:            "leader_user_id",
	CreateRequestId:         "create_request_id",
	Status:                  "status",
	Visible:                 "visible",
	CreatedAt:               "created_at",
	UpdatedAt:               "updated_at",
	DeletedAt:               "deleted_at",
	MemberCount:             "member_count",
	TotalContributionMeters: "total_contribution_meters",
	LastActiveAt:            "last_active_at",
	InvalidatedAt:           "invalidated_at",
	InvalidReason:           "invalid_reason",
	CreateResponseJson:      "create_response_json",
}

// NewTransportTeamDao creates and returns a new DAO object for table data access.
func NewTransportTeamDao(handlers ...gdb.ModelHandler) *TransportTeamDao {
	return &TransportTeamDao{
		group:    "default",
		table:    "plugin_sicau_niu_transport_team",
		columns:  transportTeamColumns,
		handlers: handlers,
	}
}

// DB retrieves and returns the underlying raw database management object of the current DAO.
func (dao *TransportTeamDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of the current DAO.
func (dao *TransportTeamDao) Table() string {
	return dao.table
}

// Columns returns all column names of the current DAO.
func (dao *TransportTeamDao) Columns() TransportTeamColumns {
	return dao.columns
}

// Group returns the database configuration group name of the current DAO.
func (dao *TransportTeamDao) Group() string {
	return dao.group
}

// Ctx creates and returns a Model for the current DAO. It automatically sets the context for the current operation.
func (dao *TransportTeamDao) Ctx(ctx context.Context) *gdb.Model {
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
func (dao *TransportTeamDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
