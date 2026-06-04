// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"time"

	"github.com/gogf/gf/v2/frame/g"
)

// Quote is the golang structure of table plugin_sicau_niu_quote for DAO operations like Where/Data.
type Quote struct {
	g.Meta    `orm:"table:plugin_sicau_niu_quote, do:true"`
	Id        any        //
	Content   any        // Quote text
	Enabled   any        // Whether the quote participates in random playback: 1 enabled, 0 disabled
	CreatedAt *time.Time // Creation time
	UpdatedAt *time.Time // Update time
	DeletedAt *time.Time // Soft-delete time, NULL means active
}
