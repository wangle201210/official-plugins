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
	AccountId int64      `json:"accountId" orm:"account_id" description:""`
	TableName string     `json:"tableName" orm:"table_name" description:""`
	Action    string     `json:"action"    orm:"action"     description:""`
	DataOld   string     `json:"dataOld"   orm:"data_old"   description:""`
	DataNew   string     `json:"dataNew"   orm:"data_new"   description:""`
	ErrMsg    string     `json:"errMsg"    orm:"err_msg"    description:""`
	ErrNumber int        `json:"errNumber" orm:"err_number" description:""`
	CreatedAt *time.Time `json:"createdAt" orm:"created_at" description:""`
	UpdatedAt *time.Time `json:"updatedAt" orm:"updated_at" description:""`
	DeletedAt *time.Time `json:"deletedAt" orm:"deleted_at" description:""`
	CreateBy  int64      `json:"createBy"  orm:"create_by"  description:""`
	UpdateBy  int64      `json:"updateBy"  orm:"update_by"  description:""`
}
