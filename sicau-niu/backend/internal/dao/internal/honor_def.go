// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// HonorDefDao is the data access object for the table plugin_sicau_niu_honor_def.
type HonorDefDao struct {
	table    string             // table is the underlying table name of the DAO.
	group    string             // group is the database configuration group name of the current DAO.
	columns  HonorDefColumns    // columns contains all the column names of Table for convenient usage.
	handlers []gdb.ModelHandler // handlers for customized model modification.
}

// HonorDefColumns defines and stores column names for the table plugin_sicau_niu_honor_def.
type HonorDefColumns struct {
	Id         string //
	HonorType  string // Honor type: badge, avatar_frame, certificate
	Code       string // Honor unique code among active rows
	Name       string //
	UnlockType string // Unlock rule: participation, feed_count, activation_count, category_complete, full_complete
	Threshold  string // Threshold for count-based unlock rules
	Category   string // Card category for category_complete unlock rule
	ImagePath  string // Honor image/template storage path
	Sort       string //
	CreatedAt  string //
	UpdatedAt  string //
	DeletedAt  string // Soft-delete time, NULL means active
}

// honorDefColumns holds the columns for the table plugin_sicau_niu_honor_def.
var honorDefColumns = HonorDefColumns{
	Id:         "id",
	HonorType:  "honor_type",
	Code:       "code",
	Name:       "name",
	UnlockType: "unlock_type",
	Threshold:  "threshold",
	Category:   "category",
	ImagePath:  "image_path",
	Sort:       "sort",
	CreatedAt:  "created_at",
	UpdatedAt:  "updated_at",
	DeletedAt:  "deleted_at",
}

// NewHonorDefDao creates and returns a new DAO object for table data access.
func NewHonorDefDao(handlers ...gdb.ModelHandler) *HonorDefDao {
	return &HonorDefDao{
		group:    "default",
		table:    "plugin_sicau_niu_honor_def",
		columns:  honorDefColumns,
		handlers: handlers,
	}
}

// DB retrieves and returns the underlying raw database management object of the current DAO.
func (dao *HonorDefDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of the current DAO.
func (dao *HonorDefDao) Table() string {
	return dao.table
}

// Columns returns all column names of the current DAO.
func (dao *HonorDefDao) Columns() HonorDefColumns {
	return dao.columns
}

// Group returns the database configuration group name of the current DAO.
func (dao *HonorDefDao) Group() string {
	return dao.group
}

// Ctx creates and returns a Model for the current DAO. It automatically sets the context for the current operation.
func (dao *HonorDefDao) Ctx(ctx context.Context) *gdb.Model {
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
func (dao *HonorDefDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
