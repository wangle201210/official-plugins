// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

import (
	"time"
)

// OauthToken is the golang structure for table oauth_token.
type OauthToken struct {
	Id        int64      `json:"id"        orm:"id"         description:""`
	TenantId  int        `json:"tenantId"  orm:"tenant_id"  description:""`
	ExpiredAt *time.Time `json:"expiredAt" orm:"expired_at" description:""`
	Code      string     `json:"code"      orm:"code"       description:""`
	Access    string     `json:"access"    orm:"access"     description:""`
	Refresh   string     `json:"refresh"   orm:"refresh"    description:""`
	Data      string     `json:"data"      orm:"data"       description:""`
	CreatedBy int64      `json:"createdBy" orm:"created_by" description:""`
	UpdatedBy int64      `json:"updatedBy" orm:"updated_by" description:""`
	CreatedAt *time.Time `json:"createdAt" orm:"created_at" description:""`
	UpdatedAt *time.Time `json:"updatedAt" orm:"updated_at" description:""`
	DeletedAt *time.Time `json:"deletedAt" orm:"deleted_at" description:""`
}
