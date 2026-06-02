// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

import (
	"time"
)

// AccountActiveLog is the golang structure for table account_active_log.
type AccountActiveLog struct {
	Id        int64      `json:"id"        orm:"id"         description:""`
	Number    string     `json:"number"    orm:"number"     description:""`
	Phone     string     `json:"phone"     orm:"phone"      description:""`
	Wechat    string     `json:"wechat"    orm:"wechat"     description:""`
	CreatedAt *time.Time `json:"createdAt" orm:"created_at" description:""`
	Type      int64      `json:"type"      orm:"type"       description:""`
}
