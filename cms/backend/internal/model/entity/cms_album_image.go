// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

import (
	"github.com/gogf/gf/v2/os/gtime"
)

// CmsAlbumImage is the golang structure for table cms_album_image.
type CmsAlbumImage struct {
	Id        int64       `json:"id"        orm:"id"         description:"Image ID"`
	AlbumId   int64       `json:"albumId"   orm:"album_id"   description:"Album ID"`
	Url       string      `json:"url"       orm:"url"        description:"Image URL"`
	Title     string      `json:"title"     orm:"title"      description:"Image title"`
	Sort      int         `json:"sort"      orm:"sort"       description:"Display order"`
	CreatedAt *gtime.Time `json:"createdAt" orm:"created_at" description:"Creation time"`
	UpdatedAt *gtime.Time `json:"updatedAt" orm:"updated_at" description:"Update time"`
}
