// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

import (
	"time"
)

// Iron is the golang structure for table iron.
type Iron struct {
	Id        int64      `json:"id"        orm:"id"         description:""`
	Code      string     `json:"code"      orm:"code"       description:"Iron-cow device identifier, unique among active rows"`
	Name      string     `json:"name"      orm:"name"       description:"Iron-cow display name"`
	LastLat   float64    `json:"lastLat"   orm:"last_lat"   description:"Latest GPS latitude pulled by the bonus flow (C4)"`
	LastLng   float64    `json:"lastLng"   orm:"last_lng"   description:"Latest GPS longitude pulled by the bonus flow (C4)"`
	LocatedAt *time.Time `json:"locatedAt" orm:"located_at" description:"Latest location pull time"`
	Remark    string     `json:"remark"    orm:"remark"     description:"Iron-cow remark"`
	CreatedAt *time.Time `json:"createdAt" orm:"created_at" description:"Creation time"`
	UpdatedAt *time.Time `json:"updatedAt" orm:"updated_at" description:"Update time"`
	DeletedAt *time.Time `json:"deletedAt" orm:"deleted_at" description:"Soft-delete time, NULL means active"`
}
