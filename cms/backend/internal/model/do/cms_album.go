// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"
)

// CmsAlbum is the golang structure of table plugin_cms_album for DAO operations like Where/Data.
type CmsAlbum struct {
	g.Meta      `orm:"table:plugin_cms_album, do:true"`
	Id          any         // Album ID
	CategoryId  any         // Category ID
	Name        any         // Album name
	Cover       any         // Cover image URL
	Description any         // Album description
	Sort        any         // Display order
	Status      any         // Status: 0=disabled, 1=enabled
	CreatedBy   any         // Creator user ID
	UpdatedBy   any         // Updater user ID
	CreatedAt   *gtime.Time // Creation time
	UpdatedAt   *gtime.Time // Update time
	DeletedAt   *gtime.Time // Deletion time
}
