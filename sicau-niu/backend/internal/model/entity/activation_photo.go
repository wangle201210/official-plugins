// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

import (
	"time"
)

// ActivationPhoto is the golang structure for table activation_photo.
type ActivationPhoto struct {
	Id           int64      `json:"id"           orm:"id"            description:""`
	Token        string     `json:"token"        orm:"token"         description:"Opaque unguessable player-facing photo identifier"`
	UserId       int64      `json:"userId"       orm:"user_id"       description:"Owning player ID"`
	RequestId    string     `json:"requestId"    orm:"request_id"    description:"Player-scoped idempotency key"`
	ActivityDate string     `json:"activityDate" orm:"activity_date" description:"Beijing natural-day key"`
	DailySlot    int        `json:"dailySlot"    orm:"daily_slot"    description:"Bounded successful upload slot 1..10 within the activity date"`
	ObjectPath   string     `json:"objectPath"   orm:"object_path"   description:"Plugin-private logical object storage path"`
	OriginalName string     `json:"originalName" orm:"original_name" description:""`
	ContentType  string     `json:"contentType"  orm:"content_type"  description:""`
	SizeBytes    int64      `json:"sizeBytes"    orm:"size_bytes"    description:""`
	UsedAt       *time.Time `json:"usedAt"       orm:"used_at"       description:"Successful activation consumption time; NULL means unused"`
	CreatedAt    *time.Time `json:"createdAt"    orm:"created_at"    description:""`
	UpdatedAt    *time.Time `json:"updatedAt"    orm:"updated_at"    description:""`
}
