// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"
)

// CmsArticleTag is the golang structure of table plugin_cms_article_tag for DAO operations like Where/Data.
type CmsArticleTag struct {
	g.Meta    `orm:"table:plugin_cms_article_tag, do:true"`
	Id        any         // Tag ID
	Name      any         // Tag name
	Slug      any         // Tag slug
	Sort      any         // Display order
	Status    any         // Status: 0=disabled, 1=enabled
	CreatedBy any         // Creator user ID
	UpdatedBy any         // Updater user ID
	CreatedAt *gtime.Time // Creation time
	UpdatedAt *gtime.Time // Update time
	DeletedAt *gtime.Time // Deletion time
}
