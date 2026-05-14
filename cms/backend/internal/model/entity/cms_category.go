// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

import (
	"github.com/gogf/gf/v2/os/gtime"
)

// CmsCategory is the golang structure for table cms_category.
type CmsCategory struct {
	Id              int64       `json:"id"              orm:"id"               description:"Category ID"`
	ParentId        int64       `json:"parentId"        orm:"parent_id"        description:"Parent category ID"`
	Code            string      `json:"code"            orm:"code"             description:"Stable category code"`
	Name            string      `json:"name"            orm:"name"             description:"Category name"`
	Type            int         `json:"type"            orm:"type"             description:"Category type: 1=list, 2=single page, 3=external link"`
	Path            string      `json:"path"            orm:"path"             description:"Public category path"`
	ListTemplate    string      `json:"listTemplate"    orm:"list_template"    description:"Public list template file"`
	ContentTemplate string      `json:"contentTemplate" orm:"content_template" description:"Public content/detail template file"`
	Cover           string      `json:"cover"           orm:"cover"            description:"Category cover image URL"`
	Outlink         string      `json:"outlink"         orm:"outlink"          description:"External link URL"`
	Title           string      `json:"title"           orm:"title"            description:"SEO title"`
	Keywords        string      `json:"keywords"        orm:"keywords"         description:"SEO keywords"`
	Description     string      `json:"description"     orm:"description"      description:"SEO description"`
	Sort            int         `json:"sort"            orm:"sort"             description:"Display order"`
	Status          int         `json:"status"          orm:"status"           description:"Status: 0=disabled, 1=enabled"`
	CreatedBy       int64       `json:"createdBy"       orm:"created_by"       description:"Creator user ID"`
	UpdatedBy       int64       `json:"updatedBy"       orm:"updated_by"       description:"Updater user ID"`
	CreatedAt       *gtime.Time `json:"createdAt"       orm:"created_at"       description:"Creation time"`
	UpdatedAt       *gtime.Time `json:"updatedAt"       orm:"updated_at"       description:"Update time"`
	DeletedAt       *gtime.Time `json:"deletedAt"       orm:"deleted_at"       description:"Deletion time"`
}
