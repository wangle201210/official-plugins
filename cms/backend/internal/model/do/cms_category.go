// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"
)

// CmsCategory is the golang structure of table plugin_cms_category for DAO operations like Where/Data.
type CmsCategory struct {
	g.Meta          `orm:"table:plugin_cms_category, do:true"`
	Id              any         // Category ID
	ParentId        any         // Parent category ID
	Code            any         // Stable category code
	Name            any         // Category name
	Type            any         // Category type: 1=list, 2=single page, 3=external link
	Path            any         // Public category path
	ListTemplate    any         // Public list template file
	ContentTemplate any         // Public content/detail template file
	Cover           any         // Category cover image URL
	Outlink         any         // External link URL
	Title           any         // SEO title
	Keywords        any         // SEO keywords
	Description     any         // SEO description
	Sort            any         // Display order
	Status          any         // Status: 0=disabled, 1=enabled
	CreatedBy       any         // Creator user ID
	UpdatedBy       any         // Updater user ID
	CreatedAt       *gtime.Time // Creation time
	UpdatedAt       *gtime.Time // Update time
	DeletedAt       *gtime.Time // Deletion time
}
