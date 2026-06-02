// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

import (
	"time"
)

// Sms is the golang structure for table sms.
type Sms struct {
	Id        int64      `json:"id"        orm:"id"         description:""`
	Phone     string     `json:"phone"     orm:"phone"      description:""`
	Type      string     `json:"type"      orm:"type"       description:""`
	Content   string     `json:"content"   orm:"content"    description:""`
	Status    int64      `json:"status"    orm:"status"     description:""`
	RespMsg   string     `json:"respMsg"   orm:"resp_msg"   description:""`
	CreatedAt *time.Time `json:"createdAt" orm:"created_at" description:""`
	UpdatedAt *time.Time `json:"updatedAt" orm:"updated_at" description:""`
	DeletedAt *time.Time `json:"deletedAt" orm:"deleted_at" description:""`
	CreateBy  int64      `json:"createBy"  orm:"create_by"  description:""`
	UpdateBy  int64      `json:"updateBy"  orm:"update_by"  description:""`
}
