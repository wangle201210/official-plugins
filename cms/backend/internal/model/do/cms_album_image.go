// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"
)

// CmsAlbumImage is the golang structure of table plugin_cms_album_image for DAO operations like Where/Data.
type CmsAlbumImage struct {
	g.Meta    `orm:"table:plugin_cms_album_image, do:true"`
	Id        any         // Image ID
	AlbumId   any         // Album ID
	Url       any         // Image URL
	Title     any         // Image title
	Sort      any         // Display order
	CreatedAt *gtime.Time // Creation time
	UpdatedAt *gtime.Time // Update time
}
