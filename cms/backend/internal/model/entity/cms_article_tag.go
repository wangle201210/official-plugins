// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

import (
	"github.com/gogf/gf/v2/os/gtime"
)

// CmsArticleTag is the golang structure for table cms_article_tag.
type CmsArticleTag struct {
	Id        int64       `json:"id"        orm:"id"         description:"Tag ID"`
	Name      string      `json:"name"      orm:"name"       description:"Tag name"`
	Slug      string      `json:"slug"      orm:"slug"       description:"Tag slug"`
	Sort      int         `json:"sort"      orm:"sort"       description:"Display order"`
	Status    int         `json:"status"    orm:"status"     description:"Status: 0=disabled, 1=enabled"`
	CreatedBy int64       `json:"createdBy" orm:"created_by" description:"Creator user ID"`
	UpdatedBy int64       `json:"updatedBy" orm:"updated_by" description:"Updater user ID"`
	CreatedAt *gtime.Time `json:"createdAt" orm:"created_at" description:"Creation time"`
	UpdatedAt *gtime.Time `json:"updatedAt" orm:"updated_at" description:"Update time"`
	DeletedAt *gtime.Time `json:"deletedAt" orm:"deleted_at" description:"Deletion time"`
}
