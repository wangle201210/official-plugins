// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

import (
	"time"
)

// OauthLog is the golang structure for table oauth_log.
type OauthLog struct {
	Id          int64      `json:"id"          orm:"id"           description:""`
	TenantId    int        `json:"tenantId"    orm:"tenant_id"    description:""`
	UserId      int64      `json:"userId"      orm:"user_id"      description:""`
	AppId       int64      `json:"appId"       orm:"app_id"       description:""`
	RedirectUri string     `json:"redirectUri" orm:"redirect_uri" description:""`
	Scope       string     `json:"scope"       orm:"scope"        description:""`
	CreatedBy   int64      `json:"createdBy"   orm:"created_by"   description:""`
	UpdatedBy   int64      `json:"updatedBy"   orm:"updated_by"   description:""`
	CreatedAt   *time.Time `json:"createdAt"   orm:"created_at"   description:""`
	UpdatedAt   *time.Time `json:"updatedAt"   orm:"updated_at"   description:""`
	DeletedAt   *time.Time `json:"deletedAt"   orm:"deleted_at"   description:""`
}
