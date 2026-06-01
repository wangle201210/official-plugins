// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

import (
	"time"
)

// Application is the golang structure for table application.
type Application struct {
	Id          int64      `json:"id"          orm:"id"           description:""`
	TenantId    int        `json:"tenantId"    orm:"tenant_id"    description:""`
	Name        string     `json:"name"        orm:"name"         description:""`
	Alias       string     `json:"alias"       orm:"alias"        description:""`
	ClientId    string     `json:"clientId"    orm:"client_id"    description:""`
	SecretKey   string     `json:"secretKey"   orm:"secret_key"   description:""`
	AccessModel string     `json:"accessModel" orm:"access_model" description:"Application access model, for example cas/oauth/ldap"`
	Status      int        `json:"status"      orm:"status"       description:"Application status: 0=disabled, 1=enabled"`
	CallbackUrl string     `json:"callbackUrl" orm:"callback_url" description:""`
	Whitelist   string     `json:"whitelist"   orm:"whitelist"    description:""`
	CreatedBy   int64      `json:"createdBy"   orm:"created_by"   description:""`
	UpdatedBy   int64      `json:"updatedBy"   orm:"updated_by"   description:""`
	CreatedAt   *time.Time `json:"createdAt"   orm:"created_at"   description:""`
	UpdatedAt   *time.Time `json:"updatedAt"   orm:"updated_at"   description:""`
	DeletedAt   *time.Time `json:"deletedAt"   orm:"deleted_at"   description:""`
}
