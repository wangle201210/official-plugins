// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

import (
	"time"
)

// AccountAppRole is the golang structure for table account_app_role.
type AccountAppRole struct {
	Id                 int64      `json:"id"                 orm:"id"                   description:""`
	TenantId           int        `json:"tenantId"           orm:"tenant_id"            description:""`
	GiveAccountId      int64      `json:"giveAccountId"      orm:"give_account_id"      description:""`
	EmpoweredAccountId int64      `json:"empoweredAccountId" orm:"empowered_account_id" description:""`
	AppId              int64      `json:"appId"              orm:"app_id"               description:""`
	ExpireAt           *time.Time `json:"expireAt"           orm:"expire_at"            description:""`
	CreatedBy          int64      `json:"createdBy"          orm:"created_by"           description:""`
	UpdatedBy          int64      `json:"updatedBy"          orm:"updated_by"           description:""`
	CreatedAt          *time.Time `json:"createdAt"          orm:"created_at"           description:""`
	UpdatedAt          *time.Time `json:"updatedAt"          orm:"updated_at"           description:""`
	DeletedAt          *time.Time `json:"deletedAt"          orm:"deleted_at"           description:""`
}
