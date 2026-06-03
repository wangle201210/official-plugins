// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

import (
	"time"
)

// Quote is the golang structure for table quote.
type Quote struct {
	Id        int64      `json:"id"        orm:"id"         description:""`
	Content   string     `json:"content"   orm:"content"    description:"Quote text"`
	Enabled   int        `json:"enabled"   orm:"enabled"    description:"Whether the quote participates in random playback: 1 enabled, 0 disabled"`
	CreatedAt *time.Time `json:"createdAt" orm:"created_at" description:"Creation time"`
	UpdatedAt *time.Time `json:"updatedAt" orm:"updated_at" description:"Update time"`
	DeletedAt *time.Time `json:"deletedAt" orm:"deleted_at" description:"Soft-delete time, NULL means active"`
}
