// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"
)

// CmsProduct is the golang structure of table plugin_cms_product for DAO operations like Where/Data.
type CmsProduct struct {
	g.Meta      `orm:"table:plugin_cms_product, do:true"`
	Id          any         // Product ID
	CategoryId  any         // Category ID
	Name        any         // Product name
	Slug        any         // Public URL slug
	Summary     any         // Product summary
	Cover       any         // Cover image URL
	Gallery     any         // Gallery image URL list as JSON array text
	Price       any         // Display price text
	Spec        any         // Specification summary
	Content     any         // Product detail HTML
	Keywords    any         // SEO keywords
	Description any         // SEO description
	Sort        any         // Display order
	Status      any         // Status: 0=draft, 1=published
	IsTop       any         // Top flag: 0=no, 1=yes
	IsRecommend any         // Recommend flag: 0=no, 1=yes
	Views       any         // View count
	PublishedAt *gtime.Time // Publication time
	CreatedBy   any         // Creator user ID
	UpdatedBy   any         // Updater user ID
	CreatedAt   *gtime.Time // Creation time
	UpdatedAt   *gtime.Time // Update time
	DeletedAt   *gtime.Time // Deletion time
}
