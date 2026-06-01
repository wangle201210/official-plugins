// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// AccountDao is the data access object for the table plugin_linapro_uidentity_cas_account.
type AccountDao struct {
	table    string             // table is the underlying table name of the DAO.
	group    string             // group is the database configuration group name of the current DAO.
	columns  AccountColumns     // columns contains all the column names of Table for convenient usage.
	handlers []gdb.ModelHandler // handlers for customized model modification.
}

// AccountColumns defines and stores column names for the table plugin_linapro_uidentity_cas_account.
type AccountColumns struct {
	Id                string //
	TenantId          string // Owning tenant ID, 0 means platform
	Number            string // Stable account number
	Name              string // Account display name
	Phone             string // Mobile phone number
	PasswordHash      string // Password hash managed by the plugin
	EffectAt          string //
	ExpireAt          string //
	PasswordUpdatedAt string //
	PassLevel         string // Password strength level: 0=invalid, higher is stronger
	ContainerId       string // Container ID
	UnitId            string // Primary unit ID
	Status            string // Account status: 0=not active, 1=normal, 2=locked
	CreatedBy         string //
	UpdatedBy         string //
	CreatedAt         string //
	UpdatedAt         string //
	DeletedAt         string //
}

// accountColumns holds the columns for the table plugin_linapro_uidentity_cas_account.
var accountColumns = AccountColumns{
	Id:                "id",
	TenantId:          "tenant_id",
	Number:            "number",
	Name:              "name",
	Phone:             "phone",
	PasswordHash:      "password_hash",
	EffectAt:          "effect_at",
	ExpireAt:          "expire_at",
	PasswordUpdatedAt: "password_updated_at",
	PassLevel:         "pass_level",
	ContainerId:       "container_id",
	UnitId:            "unit_id",
	Status:            "status",
	CreatedBy:         "created_by",
	UpdatedBy:         "updated_by",
	CreatedAt:         "created_at",
	UpdatedAt:         "updated_at",
	DeletedAt:         "deleted_at",
}

// NewAccountDao creates and returns a new DAO object for table data access.
func NewAccountDao(handlers ...gdb.ModelHandler) *AccountDao {
	return &AccountDao{
		group:    "default",
		table:    "plugin_linapro_uidentity_cas_account",
		columns:  accountColumns,
		handlers: handlers,
	}
}

// DB retrieves and returns the underlying raw database management object of the current DAO.
func (dao *AccountDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of the current DAO.
func (dao *AccountDao) Table() string {
	return dao.table
}

// Columns returns all column names of the current DAO.
func (dao *AccountDao) Columns() AccountColumns {
	return dao.columns
}

// Group returns the database configuration group name of the current DAO.
func (dao *AccountDao) Group() string {
	return dao.group
}

// Ctx creates and returns a Model for the current DAO. It automatically sets the context for the current operation.
func (dao *AccountDao) Ctx(ctx context.Context) *gdb.Model {
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
func (dao *AccountDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
