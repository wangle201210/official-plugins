// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"
)

// CmsArticle is the golang structure of table plugin_cms_article for DAO operations like Where/Data.
type CmsArticle struct {
	g.Meta      `orm:"table:plugin_cms_article, do:true"`
	Id          any         // Article ID
	CategoryId  any         // Category ID
	Title       any         // Article title
	Subtitle    any         // Article subtitle
	Slug        any         // Public URL slug
	Summary     any         // Article summary
	Cover       any         // Cover image URL
	Author      any         // Author name
	Source      any         // Content source
	Content     any         // Article body HTML
	Tags        any         // Comma-separated tag names
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
