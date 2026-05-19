// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"time"

	"github.com/gogf/gf/v2/frame/g"
)

// Record is the golang structure of table plugin_linapro_demo_source_record for DAO operations like Where/Data.
type Record struct {
	g.Meta         `orm:"table:plugin_linapro_demo_source_record, do:true"`
	Id             any        // Primary key ID
	TenantId       any        // Owning tenant ID, 0 means PLATFORM
	Title          any        // Record title
	Content        any        // Record content
	AttachmentName any        // Original attachment file name
	AttachmentPath any        // Relative attachment storage path
	CreatedAt      *time.Time // Creation time
	UpdatedAt      *time.Time // Update time
}
