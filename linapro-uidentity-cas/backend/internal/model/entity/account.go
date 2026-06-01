// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

import (
	"time"
)

// Account is the golang structure for table account.
type Account struct {
	Id                int64      `json:"id"                orm:"id"                  description:""`
	TenantId          int        `json:"tenantId"          orm:"tenant_id"           description:"Owning tenant ID, 0 means platform"`
	Number            string     `json:"number"            orm:"number"              description:"Stable account number"`
	Name              string     `json:"name"              orm:"name"                description:"Account display name"`
	Phone             string     `json:"phone"             orm:"phone"               description:"Mobile phone number"`
	PasswordHash      string     `json:"passwordHash"      orm:"password_hash"       description:"Password hash managed by the plugin"`
	EffectAt          *time.Time `json:"effectAt"          orm:"effect_at"           description:""`
	ExpireAt          *time.Time `json:"expireAt"          orm:"expire_at"           description:""`
	PasswordUpdatedAt *time.Time `json:"passwordUpdatedAt" orm:"password_updated_at" description:""`
	PassLevel         int        `json:"passLevel"         orm:"pass_level"          description:"Password strength level: 0=invalid, higher is stronger"`
	ContainerId       int64      `json:"containerId"       orm:"container_id"        description:"Container ID"`
	UnitId            int64      `json:"unitId"            orm:"unit_id"             description:"Primary unit ID"`
	Status            int        `json:"status"            orm:"status"              description:"Account status: 0=not active, 1=normal, 2=locked"`
	CreatedBy         int64      `json:"createdBy"         orm:"created_by"          description:""`
	UpdatedBy         int64      `json:"updatedBy"         orm:"updated_by"          description:""`
	CreatedAt         *time.Time `json:"createdAt"         orm:"created_at"          description:""`
	UpdatedAt         *time.Time `json:"updatedAt"         orm:"updated_at"          description:""`
	DeletedAt         *time.Time `json:"deletedAt"         orm:"deleted_at"          description:""`
}
