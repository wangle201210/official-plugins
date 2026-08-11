// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// ActivationPhotoDao is the data access object for the table plugin_sicau_niu_activation_photo.
type ActivationPhotoDao struct {
	table    string                 // table is the underlying table name of the DAO.
	group    string                 // group is the database configuration group name of the current DAO.
	columns  ActivationPhotoColumns // columns contains all the column names of Table for convenient usage.
	handlers []gdb.ModelHandler     // handlers for customized model modification.
}

// ActivationPhotoColumns defines and stores column names for the table plugin_sicau_niu_activation_photo.
type ActivationPhotoColumns struct {
	Id           string //
	Token        string // Opaque unguessable player-facing photo identifier
	UserId       string // Owning player ID
	RequestId    string // Player-scoped idempotency key
	ActivityDate string // Beijing natural-day key
	DailySlot    string // Bounded successful upload slot 1..10 within the activity date
	ObjectPath   string // Plugin-private logical object storage path
	OriginalName string //
	ContentType  string //
	SizeBytes    string //
	UsedAt       string // Successful activation consumption time; NULL means unused
	CreatedAt    string //
	UpdatedAt    string //
}

// activationPhotoColumns holds the columns for the table plugin_sicau_niu_activation_photo.
var activationPhotoColumns = ActivationPhotoColumns{
	Id:           "id",
	Token:        "token",
	UserId:       "user_id",
	RequestId:    "request_id",
	ActivityDate: "activity_date",
	DailySlot:    "daily_slot",
	ObjectPath:   "object_path",
	OriginalName: "original_name",
	ContentType:  "content_type",
	SizeBytes:    "size_bytes",
	UsedAt:       "used_at",
	CreatedAt:    "created_at",
	UpdatedAt:    "updated_at",
}

// NewActivationPhotoDao creates and returns a new DAO object for table data access.
func NewActivationPhotoDao(handlers ...gdb.ModelHandler) *ActivationPhotoDao {
	return &ActivationPhotoDao{
		group:    "default",
		table:    "plugin_sicau_niu_activation_photo",
		columns:  activationPhotoColumns,
		handlers: handlers,
	}
}

// DB retrieves and returns the underlying raw database management object of the current DAO.
func (dao *ActivationPhotoDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of the current DAO.
func (dao *ActivationPhotoDao) Table() string {
	return dao.table
}

// Columns returns all column names of the current DAO.
func (dao *ActivationPhotoDao) Columns() ActivationPhotoColumns {
	return dao.columns
}

// Group returns the database configuration group name of the current DAO.
func (dao *ActivationPhotoDao) Group() string {
	return dao.group
}

// Ctx creates and returns a Model for the current DAO. It automatically sets the context for the current operation.
func (dao *ActivationPhotoDao) Ctx(ctx context.Context) *gdb.Model {
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
func (dao *ActivationPhotoDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
