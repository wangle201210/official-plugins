// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

import (
	"time"
)

// College is the golang structure for table college.
type College struct {
	Id        int64      `json:"id"        orm:"id"         description:"Primary key ID"`
	Name      string     `json:"name"      orm:"name"       description:"College name"`
	Sort      int        `json:"sort"      orm:"sort"       description:"Display sort order, smaller first"`
	CreatedAt *time.Time `json:"createdAt" orm:"created_at" description:"Creation time"`
	UpdatedAt *time.Time `json:"updatedAt" orm:"updated_at" description:"Update time"`
	DeletedAt *time.Time `json:"deletedAt" orm:"deleted_at" description:"Soft-delete time, NULL means active"`
}
