// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

import (
	"github.com/gogf/gf/v2/os/gtime"
)

// CmsSlide is the golang structure for table cms_slide.
type CmsSlide struct {
	Id        int64       `json:"id"        orm:"id"         description:"Slide ID"`
	GroupCode string      `json:"groupCode" orm:"group_code" description:"Display group code"`
	Title     string      `json:"title"     orm:"title"      description:"Slide title"`
	Subtitle  string      `json:"subtitle"  orm:"subtitle"   description:"Slide subtitle"`
	Image     string      `json:"image"     orm:"image"      description:"Slide image URL"`
	Link      string      `json:"link"      orm:"link"       description:"Click target URL"`
	Sort      int         `json:"sort"      orm:"sort"       description:"Display order"`
	Status    int         `json:"status"    orm:"status"     description:"Status: 0=disabled, 1=enabled"`
	CreatedBy int64       `json:"createdBy" orm:"created_by" description:"Creator user ID"`
	UpdatedBy int64       `json:"updatedBy" orm:"updated_by" description:"Updater user ID"`
	CreatedAt *gtime.Time `json:"createdAt" orm:"created_at" description:"Creation time"`
	UpdatedAt *gtime.Time `json:"updatedAt" orm:"updated_at" description:"Update time"`
	DeletedAt *gtime.Time `json:"deletedAt" orm:"deleted_at" description:"Deletion time"`
}
