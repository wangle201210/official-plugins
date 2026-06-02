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
	GiveAccountId      int64      `json:"giveAccountId"      orm:"give_account_id"      description:""`
	EmpoweredAccountId int64      `json:"empoweredAccountId" orm:"empowered_account_id" description:""`
	AppId              int64      `json:"appId"              orm:"app_id"               description:""`
	ExpireAt           *time.Time `json:"expireAt"           orm:"expire_at"            description:""`
	CreatedAt          *time.Time `json:"createdAt"          orm:"created_at"           description:""`
	UpdatedAt          *time.Time `json:"updatedAt"          orm:"updated_at"           description:""`
	DeletedAt          *time.Time `json:"deletedAt"          orm:"deleted_at"           description:""`
	CreateBy           int64      `json:"createBy"           orm:"create_by"            description:""`
	UpdateBy           int64      `json:"updateBy"           orm:"update_by"            description:""`
}
