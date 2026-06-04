// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"
)

// CmsLink is the golang structure of table plugin_cms_link for DAO operations like Where/Data.
type CmsLink struct {
	g.Meta    `orm:"table:plugin_cms_link, do:true"`
	Id        any         // Link ID
	GroupCode any         // Display group code
	Name      any         // Link name
	Url       any         // Link URL
	Logo      any         // Logo URL
	Sort      any         // Display order
	Status    any         // Status: 0=disabled, 1=enabled
	CreatedBy any         // Creator user ID
	UpdatedBy any         // Updater user ID
	CreatedAt *gtime.Time // Creation time
	UpdatedAt *gtime.Time // Update time
	DeletedAt *gtime.Time // Deletion time
}
