// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// TransportSessionDao is the data access object for the table plugin_sicau_niu_transport_session.
type TransportSessionDao struct {
	table    string                  // table is the underlying table name of the DAO.
	group    string                  // group is the database configuration group name of the current DAO.
	columns  TransportSessionColumns // columns contains all the column names of Table for convenient usage.
	handlers []gdb.ModelHandler      // handlers for customized model modification.
}

// TransportSessionColumns defines and stores column names for the table plugin_sicau_niu_transport_session.
type TransportSessionColumns struct {
	Id              string //
	TeamId          string //
	IronId          string //
	StartedByUserId string //
	StartRequestId  string //
	EndedByUserId   string //
	EndRequestId    string //
	Status          string // Session status: active, idle_timeout, ended
	StartedAt       string //
	LastActiveAt    string //
	EndedAt         string //
	MovedMeters     string //
	LastLat         string //
	LastLng         string //
	CreatedAt       string //
	UpdatedAt       string //
}

// transportSessionColumns holds the columns for the table plugin_sicau_niu_transport_session.
var transportSessionColumns = TransportSessionColumns{
	Id:              "id",
	TeamId:          "team_id",
	IronId:          "iron_id",
	StartedByUserId: "started_by_user_id",
	StartRequestId:  "start_request_id",
	EndedByUserId:   "ended_by_user_id",
	EndRequestId:    "end_request_id",
	Status:          "status",
	StartedAt:       "started_at",
	LastActiveAt:    "last_active_at",
	EndedAt:         "ended_at",
	MovedMeters:     "moved_meters",
	LastLat:         "last_lat",
	LastLng:         "last_lng",
	CreatedAt:       "created_at",
	UpdatedAt:       "updated_at",
}

// NewTransportSessionDao creates and returns a new DAO object for table data access.
func NewTransportSessionDao(handlers ...gdb.ModelHandler) *TransportSessionDao {
	return &TransportSessionDao{
		group:    "default",
		table:    "plugin_sicau_niu_transport_session",
		columns:  transportSessionColumns,
		handlers: handlers,
	}
}

// DB retrieves and returns the underlying raw database management object of the current DAO.
func (dao *TransportSessionDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of the current DAO.
func (dao *TransportSessionDao) Table() string {
	return dao.table
}

// Columns returns all column names of the current DAO.
func (dao *TransportSessionDao) Columns() TransportSessionColumns {
	return dao.columns
}

// Group returns the database configuration group name of the current DAO.
func (dao *TransportSessionDao) Group() string {
	return dao.group
}

// Ctx creates and returns a Model for the current DAO. It automatically sets the context for the current operation.
func (dao *TransportSessionDao) Ctx(ctx context.Context) *gdb.Model {
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
func (dao *TransportSessionDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
