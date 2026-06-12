// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// ActivationAttemptDao is the data access object for the table plugin_sicau_niu_activation_attempt.
type ActivationAttemptDao struct {
	table    string                   // table is the underlying table name of the DAO.
	group    string                   // group is the database configuration group name of the current DAO.
	columns  ActivationAttemptColumns // columns contains all the column names of Table for convenient usage.
	handlers []gdb.ModelHandler       // handlers for customized model modification.
}

// ActivationAttemptColumns defines and stores column names for the table plugin_sicau_niu_activation_attempt.
type ActivationAttemptColumns struct {
	Id           string //
	UserId       string // Player ID that submitted the photo check-in
	NiuId        string // Activated cattle ID when result is success, otherwise 0
	NearestNiuId string // Nearest visible inactive cattle candidate ID when available
	Result       string // Attempt result: success, no_nearby, out_of_range, speed_anomaly
	Lat          string // Player reported GPS latitude
	Lng          string // Player reported GPS longitude
	DistanceM    string // Distance in meters to nearest candidate, 0 when unavailable
	ThresholdM   string // LBS activation threshold in meters used for the attempt
	PhotoPath    string // Uploaded photo storage path, evidence only
	AttemptedAt  string // Attempt time
	CreatedAt    string // Creation time
	UpdatedAt    string // Update time
	DeletedAt    string // Soft-delete time, NULL means active
}

// activationAttemptColumns holds the columns for the table plugin_sicau_niu_activation_attempt.
var activationAttemptColumns = ActivationAttemptColumns{
	Id:           "id",
	UserId:       "user_id",
	NiuId:        "niu_id",
	NearestNiuId: "nearest_niu_id",
	Result:       "result",
	Lat:          "lat",
	Lng:          "lng",
	DistanceM:    "distance_m",
	ThresholdM:   "threshold_m",
	PhotoPath:    "photo_path",
	AttemptedAt:  "attempted_at",
	CreatedAt:    "created_at",
	UpdatedAt:    "updated_at",
	DeletedAt:    "deleted_at",
}

// NewActivationAttemptDao creates and returns a new DAO object for table data access.
func NewActivationAttemptDao(handlers ...gdb.ModelHandler) *ActivationAttemptDao {
	return &ActivationAttemptDao{
		group:    "default",
		table:    "plugin_sicau_niu_activation_attempt",
		columns:  activationAttemptColumns,
		handlers: handlers,
	}
}

// DB retrieves and returns the underlying raw database management object of the current DAO.
func (dao *ActivationAttemptDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of the current DAO.
func (dao *ActivationAttemptDao) Table() string {
	return dao.table
}

// Columns returns all column names of the current DAO.
func (dao *ActivationAttemptDao) Columns() ActivationAttemptColumns {
	return dao.columns
}

// Group returns the database configuration group name of the current DAO.
func (dao *ActivationAttemptDao) Group() string {
	return dao.group
}

// Ctx creates and returns a Model for the current DAO. It automatically sets the context for the current operation.
func (dao *ActivationAttemptDao) Ctx(ctx context.Context) *gdb.Model {
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
func (dao *ActivationAttemptDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
