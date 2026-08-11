// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// MiniappConfigDao is the data access object for the table plugin_sicau_niu_miniapp_config.
type MiniappConfigDao struct {
	table    string               // table is the underlying table name of the DAO.
	group    string               // group is the database configuration group name of the current DAO.
	columns  MiniappConfigColumns // columns contains all the column names of Table for convenient usage.
	handlers []gdb.ModelHandler   // handlers for customized model modification.
}

// MiniappConfigColumns defines and stores column names for the table plugin_sicau_niu_miniapp_config.
type MiniappConfigColumns struct {
	Id                 string //
	ConfigKey          string //
	AssetsVersion      string //
	StaticAssetBaseUrl string //
	ActivityPhase      string //
	DefaultCampus      string //
	Anniversary        string //
	AnniversaryAt      string //
	Debug              string //
	CampusesJson       string // Validated campus map asset and affine calibration JSON
	CreatedAt          string //
	UpdatedAt          string //
}

// miniappConfigColumns holds the columns for the table plugin_sicau_niu_miniapp_config.
var miniappConfigColumns = MiniappConfigColumns{
	Id:                 "id",
	ConfigKey:          "config_key",
	AssetsVersion:      "assets_version",
	StaticAssetBaseUrl: "static_asset_base_url",
	ActivityPhase:      "activity_phase",
	DefaultCampus:      "default_campus",
	Anniversary:        "anniversary",
	AnniversaryAt:      "anniversary_at",
	Debug:              "debug",
	CampusesJson:       "campuses_json",
	CreatedAt:          "created_at",
	UpdatedAt:          "updated_at",
}

// NewMiniappConfigDao creates and returns a new DAO object for table data access.
func NewMiniappConfigDao(handlers ...gdb.ModelHandler) *MiniappConfigDao {
	return &MiniappConfigDao{
		group:    "default",
		table:    "plugin_sicau_niu_miniapp_config",
		columns:  miniappConfigColumns,
		handlers: handlers,
	}
}

// DB retrieves and returns the underlying raw database management object of the current DAO.
func (dao *MiniappConfigDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of the current DAO.
func (dao *MiniappConfigDao) Table() string {
	return dao.table
}

// Columns returns all column names of the current DAO.
func (dao *MiniappConfigDao) Columns() MiniappConfigColumns {
	return dao.columns
}

// Group returns the database configuration group name of the current DAO.
func (dao *MiniappConfigDao) Group() string {
	return dao.group
}

// Ctx creates and returns a Model for the current DAO. It automatically sets the context for the current operation.
func (dao *MiniappConfigDao) Ctx(ctx context.Context) *gdb.Model {
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
func (dao *MiniappConfigDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
