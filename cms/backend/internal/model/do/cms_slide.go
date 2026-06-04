// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"
)

// CmsSlide is the golang structure of table plugin_cms_slide for DAO operations like Where/Data.
type CmsSlide struct {
	g.Meta    `orm:"table:plugin_cms_slide, do:true"`
	Id        any         // Slide ID
	GroupCode any         // Display group code
	Title     any         // Slide title
	Subtitle  any         // Slide subtitle
	Image     any         // Slide image URL
	Link      any         // Click target URL
	Sort      any         // Display order
	Status    any         // Status: 0=disabled, 1=enabled
	CreatedBy any         // Creator user ID
	UpdatedBy any         // Updater user ID
	CreatedAt *gtime.Time // Creation time
	UpdatedAt *gtime.Time // Update time
	DeletedAt *gtime.Time // Deletion time
}
