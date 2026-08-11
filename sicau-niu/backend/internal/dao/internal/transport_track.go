// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// TransportTrackDao is the data access object for the table plugin_sicau_niu_transport_track.
type TransportTrackDao struct {
	table    string                // table is the underlying table name of the DAO.
	group    string                // group is the database configuration group name of the current DAO.
	columns  TransportTrackColumns // columns contains all the column names of Table for convenient usage.
	handlers []gdb.ModelHandler    // handlers for customized model modification.
}

// TransportTrackColumns defines and stores column names for the table plugin_sicau_niu_transport_track.
type TransportTrackColumns struct {
	Id             string //
	SessionId      string //
	UserId         string //
	RequestId      string //
	Lat            string //
	Lng            string //
	DistanceMeters string //
	RecordedAt     string //
	CreatedAt      string //
}

// transportTrackColumns holds the columns for the table plugin_sicau_niu_transport_track.
var transportTrackColumns = TransportTrackColumns{
	Id:             "id",
	SessionId:      "session_id",
	UserId:         "user_id",
	RequestId:      "request_id",
	Lat:            "lat",
	Lng:            "lng",
	DistanceMeters: "distance_meters",
	RecordedAt:     "recorded_at",
	CreatedAt:      "created_at",
}

// NewTransportTrackDao creates and returns a new DAO object for table data access.
func NewTransportTrackDao(handlers ...gdb.ModelHandler) *TransportTrackDao {
	return &TransportTrackDao{
		group:    "default",
		table:    "plugin_sicau_niu_transport_track",
		columns:  transportTrackColumns,
		handlers: handlers,
	}
}

// DB retrieves and returns the underlying raw database management object of the current DAO.
func (dao *TransportTrackDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of the current DAO.
func (dao *TransportTrackDao) Table() string {
	return dao.table
}

// Columns returns all column names of the current DAO.
func (dao *TransportTrackDao) Columns() TransportTrackColumns {
	return dao.columns
}

// Group returns the database configuration group name of the current DAO.
func (dao *TransportTrackDao) Group() string {
	return dao.group
}

// Ctx creates and returns a Model for the current DAO. It automatically sets the context for the current operation.
func (dao *TransportTrackDao) Ctx(ctx context.Context) *gdb.Model {
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
func (dao *TransportTrackDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
