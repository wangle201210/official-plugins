// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"
)

// CmsMessage is the golang structure of table plugin_cms_message for DAO operations like Where/Data.
type CmsMessage struct {
	g.Meta    `orm:"table:plugin_cms_message, do:true"`
	Id        any         // Message ID
	Name      any         // Visitor name
	Mobile    any         // Visitor mobile
	Email     any         // Visitor email
	Content   any         // Message content
	Reply     any         // Reply content
	Status    any         // Status: 0=pending, 1=approved, 2=rejected
	UserIp    any         // Visitor IP
	UserAgent any         // Visitor user agent
	CreatedBy any         // Creator user ID
	UpdatedBy any         // Updater user ID
	CreatedAt *gtime.Time // Creation time
	UpdatedAt *gtime.Time // Update time
	DeletedAt *gtime.Time // Deletion time
}
