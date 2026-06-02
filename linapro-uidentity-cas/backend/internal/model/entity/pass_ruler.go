// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

import (
	"time"
)

// PassRuler is the golang structure for table pass_ruler.
type PassRuler struct {
	Id             int64      `json:"id"             orm:"id"              description:""`
	Name           string     `json:"name"           orm:"name"            description:""`
	Capital        int64      `json:"capital"        orm:"capital"         description:""`
	Lower          int64      `json:"lower"          orm:"lower"           description:""`
	Number         int64      `json:"number"         orm:"number"          description:""`
	Symbol         int64      `json:"symbol"         orm:"symbol"          description:""`
	Length         int64      `json:"length"         orm:"length"          description:""`
	Interval       int64      `json:"interval"       orm:"interval"        description:""`
	IntervalStatus int        `json:"intervalStatus" orm:"interval_status" description:""`
	Status         int        `json:"status"         orm:"status"          description:""`
	CreatedAt      *time.Time `json:"createdAt"      orm:"created_at"      description:""`
	UpdatedAt      *time.Time `json:"updatedAt"      orm:"updated_at"      description:""`
	DeletedAt      *time.Time `json:"deletedAt"      orm:"deleted_at"      description:""`
	CreateBy       int64      `json:"createBy"       orm:"create_by"       description:""`
	UpdateBy       int64      `json:"updateBy"       orm:"update_by"       description:""`
}
