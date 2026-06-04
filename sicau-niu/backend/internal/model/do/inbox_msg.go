// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"time"

	"github.com/gogf/gf/v2/frame/g"
)

// InboxMsg is the golang structure of table plugin_sicau_niu_inbox_msg for DAO operations like Where/Data.
type InboxMsg struct {
	g.Meta    `orm:"table:plugin_sicau_niu_inbox_msg, do:true"`
	Id        any        //
	UserId    any        //
	MsgType   any        // Message type: stolen, gift_received
	Content   any        //
	IsRead    any        // Read flag: 1 read, 0 unread
	CreatedAt *time.Time //
	UpdatedAt *time.Time //
	DeletedAt *time.Time //
}
