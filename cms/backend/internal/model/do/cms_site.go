// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"
)

// CmsSite is the golang structure of table plugin_cms_site for DAO operations like Where/Data.
type CmsSite struct {
	g.Meta      `orm:"table:plugin_cms_site, do:true"`
	Id          any         // Site ID
	SiteKey     any         // Stable site key
	Name        any         // Site name
	Logo        any         // Site logo URL
	Weixin      any         // WeChat QR code image URL
	Domain      any         // Primary site domain
	Slogan      any         // Site slogan
	Keywords    any         // SEO keywords
	Description any         // SEO description
	Icp         any         // ICP record number
	Contact     any         // Contact person
	Phone       any         // Contact phone
	Email       any         // Contact email
	Address     any         // Contact address
	Status      any         // Status: 0=disabled, 1=enabled
	CreatedBy   any         // Creator user ID
	UpdatedBy   any         // Updater user ID
	CreatedAt   *gtime.Time // Creation time
	UpdatedAt   *gtime.Time // Update time
	DeletedAt   *gtime.Time // Deletion time
}
