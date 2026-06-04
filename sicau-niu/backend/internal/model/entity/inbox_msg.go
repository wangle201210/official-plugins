// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

import (
	"time"
)

// InboxMsg is the golang structure for table inbox_msg.
type InboxMsg struct {
	Id        int64      `json:"id"        orm:"id"         description:""`
	UserId    int64      `json:"userId"    orm:"user_id"    description:""`
	MsgType   string     `json:"msgType"   orm:"msg_type"   description:"Message type: stolen, gift_received"`
	Content   string     `json:"content"   orm:"content"    description:""`
	IsRead    int        `json:"isRead"    orm:"is_read"    description:"Read flag: 1 read, 0 unread"`
	CreatedAt *time.Time `json:"createdAt" orm:"created_at" description:""`
	UpdatedAt *time.Time `json:"updatedAt" orm:"updated_at" description:""`
	DeletedAt *time.Time `json:"deletedAt" orm:"deleted_at" description:""`
}
