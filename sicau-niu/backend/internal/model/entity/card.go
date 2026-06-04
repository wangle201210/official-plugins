// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

import (
	"time"
)

// Card is the golang structure for table card.
type Card struct {
	Id        int64      `json:"id"        orm:"id"         description:""`
	NiuId     int64      `json:"niuId"     orm:"niu_id"     description:"Owning cattle ID; one card per cattle"`
	Category  string     `json:"category"  orm:"category"   description:"Card category: person, event, research, college, spirit"`
	Title     string     `json:"title"     orm:"title"      description:"Card title"`
	Content   string     `json:"content"   orm:"content"    description:"Card content text"`
	ImagePath string     `json:"imagePath" orm:"image_path" description:"Card image storage path uploaded via host file management"`
	CreatedAt *time.Time `json:"createdAt" orm:"created_at" description:"Creation time"`
	UpdatedAt *time.Time `json:"updatedAt" orm:"updated_at" description:"Update time"`
	DeletedAt *time.Time `json:"deletedAt" orm:"deleted_at" description:"Soft-delete time, NULL means active"`
}
