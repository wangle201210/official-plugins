// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

import (
	"time"
)

// AccountChangeLog is the golang structure for table account_change_log.
type AccountChangeLog struct {
	Id        int64      `json:"id"        orm:"id"         description:""`
	TenantId  int        `json:"tenantId"  orm:"tenant_id"  description:""`
	AccountId int64      `json:"accountId" orm:"account_id" description:""`
	TableName string     `json:"tableName" orm:"table_name" description:""`
	Action    string     `json:"action"    orm:"action"     description:""`
	DataOld   string     `json:"dataOld"   orm:"data_old"   description:""`
	DataNew   string     `json:"dataNew"   orm:"data_new"   description:""`
	ErrMsg    string     `json:"errMsg"    orm:"err_msg"    description:""`
	ErrNumber string     `json:"errNumber" orm:"err_number" description:""`
	CreatedBy int64      `json:"createdBy" orm:"created_by" description:""`
	UpdatedBy int64      `json:"updatedBy" orm:"updated_by" description:""`
	CreatedAt *time.Time `json:"createdAt" orm:"created_at" description:""`
	UpdatedAt *time.Time `json:"updatedAt" orm:"updated_at" description:""`
	DeletedAt *time.Time `json:"deletedAt" orm:"deleted_at" description:""`
}
