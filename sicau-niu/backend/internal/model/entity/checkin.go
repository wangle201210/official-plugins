// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

import (
	"time"
)

// Checkin is the golang structure for table checkin.
type Checkin struct {
	Id          int64      `json:"id"          orm:"id"           description:""`
	UserId      int64      `json:"userId"      orm:"user_id"      description:""`
	CheckinDate string     `json:"checkinDate" orm:"checkin_date" description:"Check-in date YYYY-MM-DD"`
	Amount      int        `json:"amount"      orm:"amount"       description:"Granted grass amount (20-50)"`
	CreatedAt   *time.Time `json:"createdAt"   orm:"created_at"   description:""`
	UpdatedAt   *time.Time `json:"updatedAt"   orm:"updated_at"   description:""`
	DeletedAt   *time.Time `json:"deletedAt"   orm:"deleted_at"   description:""`
}
