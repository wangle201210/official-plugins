// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

import (
	"github.com/gogf/gf/v2/os/gtime"
)

// CmsMessage is the golang structure for table cms_message.
type CmsMessage struct {
	Id        int64       `json:"id"        orm:"id"         description:"Message ID"`
	Name      string      `json:"name"      orm:"name"       description:"Visitor name"`
	Mobile    string      `json:"mobile"    orm:"mobile"     description:"Visitor mobile"`
	Email     string      `json:"email"     orm:"email"      description:"Visitor email"`
	Content   string      `json:"content"   orm:"content"    description:"Message content"`
	Reply     string      `json:"reply"     orm:"reply"      description:"Reply content"`
	Status    int         `json:"status"    orm:"status"     description:"Status: 0=pending, 1=approved, 2=rejected"`
	UserIp    string      `json:"userIp"    orm:"user_ip"    description:"Visitor IP"`
	UserAgent string      `json:"userAgent" orm:"user_agent" description:"Visitor user agent"`
	CreatedBy int64       `json:"createdBy" orm:"created_by" description:"Creator user ID"`
	UpdatedBy int64       `json:"updatedBy" orm:"updated_by" description:"Updater user ID"`
	CreatedAt *gtime.Time `json:"createdAt" orm:"created_at" description:"Creation time"`
	UpdatedAt *gtime.Time `json:"updatedAt" orm:"updated_at" description:"Update time"`
	DeletedAt *gtime.Time `json:"deletedAt" orm:"deleted_at" description:"Deletion time"`
}
