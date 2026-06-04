// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

import (
	"time"
)

// HonorDef is the golang structure for table honor_def.
type HonorDef struct {
	Id         int64      `json:"id"         orm:"id"          description:""`
	HonorType  string     `json:"honorType"  orm:"honor_type"  description:"Honor type: badge, avatar_frame, certificate"`
	Code       string     `json:"code"       orm:"code"        description:"Honor unique code among active rows"`
	Name       string     `json:"name"       orm:"name"        description:""`
	UnlockType string     `json:"unlockType" orm:"unlock_type" description:"Unlock rule: participation, feed_count, activation_count, category_complete, full_complete"`
	Threshold  int        `json:"threshold"  orm:"threshold"   description:"Threshold for count-based unlock rules"`
	Category   string     `json:"category"   orm:"category"    description:"Card category for category_complete unlock rule"`
	ImagePath  string     `json:"imagePath"  orm:"image_path"  description:"Honor image/template storage path"`
	Sort       int        `json:"sort"       orm:"sort"        description:""`
	CreatedAt  *time.Time `json:"createdAt"  orm:"created_at"  description:""`
	UpdatedAt  *time.Time `json:"updatedAt"  orm:"updated_at"  description:""`
	DeletedAt  *time.Time `json:"deletedAt"  orm:"deleted_at"  description:"Soft-delete time, NULL means active"`
}
