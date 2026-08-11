// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

import (
	"time"
)

// Activation is the golang structure for table activation.
type Activation struct {
	Id           int64      `json:"id"           orm:"id"            description:""`
	UserId       int64      `json:"userId"       orm:"user_id"       description:"Activating player ID"`
	NiuId        int64      `json:"niuId"        orm:"niu_id"        description:"Activated cattle ID"`
	ActivityDate string     `json:"activityDate" orm:"activity_date" description:"Activation date YYYY-MM-DD for the per-day limit"`
	ActivatedAt  *time.Time `json:"activatedAt"  orm:"activated_at"  description:"Activation time"`
	IsFirst      int        `json:"isFirst"      orm:"is_first"      description:"Whether this is the cattle first-activation: 1 first, 0 later"`
	OrderNo      int        `json:"orderNo"      orm:"order_no"      description:"Arrival order for this cattle, starting at 1"`
	PhotoPath    string     `json:"photoPath"    orm:"photo_path"    description:"Optional activation photo storage path (evidence only, no recognition)"`
	CreatedAt    *time.Time `json:"createdAt"    orm:"created_at"    description:"Creation time"`
	UpdatedAt    *time.Time `json:"updatedAt"    orm:"updated_at"    description:"Update time"`
	DeletedAt    *time.Time `json:"deletedAt"    orm:"deleted_at"    description:"Soft-delete time, NULL means active"`
	RequestId    string     `json:"requestId"    orm:"request_id"    description:"Player-scoped idempotency key"`
	ResponseJson string     `json:"responseJson" orm:"response_json" description:"Stable successful activation response for idempotent replay"`
}
