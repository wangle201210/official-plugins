// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

import (
	"time"
)

// Applications is the golang structure for table applications.
type Applications struct {
	Id          int64      `json:"id"          orm:"id"           description:""`
	Name        string     `json:"name"        orm:"name"         description:""`
	Alias       string     `json:"alias"       orm:"alias"        description:""`
	ClientId    string     `json:"clientId"    orm:"client_id"    description:""`
	SecretKey   string     `json:"secretKey"   orm:"secret_key"   description:""`
	AccessModel string     `json:"accessModel" orm:"access_model" description:""`
	Status      int64      `json:"status"      orm:"status"       description:""`
	CallbackUrl string     `json:"callbackUrl" orm:"callback_url" description:""`
	Whitelist   string     `json:"whitelist"   orm:"whitelist"    description:""`
	CreatedAt   *time.Time `json:"createdAt"   orm:"created_at"   description:""`
	UpdatedAt   *time.Time `json:"updatedAt"   orm:"updated_at"   description:""`
	DeletedAt   *time.Time `json:"deletedAt"   orm:"deleted_at"   description:""`
	CreateBy    int64      `json:"createBy"    orm:"create_by"    description:""`
	UpdateBy    int64      `json:"updateBy"    orm:"update_by"    description:""`
}
