// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

import (
	"time"
)

// Niu is the golang structure for table niu.
type Niu struct {
	Id              int64      `json:"id"              orm:"id"               description:""`
	Code            string     `json:"code"            orm:"code"             description:"Cattle serial code, unique among active rows"`
	NiuType         string     `json:"niuType"         orm:"niu_type"         description:"Cattle type: common, special"`
	SpecialSubtype  string     `json:"specialSubtype"  orm:"special_subtype"  description:"Special subtype: college, contribution, alumni, spirit; empty for common"`
	Name            string     `json:"name"            orm:"name"             description:"Cattle name, used by special cattle"`
	CollegeId       int64      `json:"collegeId"       orm:"college_id"       description:"Linked college ID for college cattle, 0 means none"`
	Lat             float64    `json:"lat"             orm:"lat"              description:"GPS latitude anchor"`
	Lng             float64    `json:"lng"             orm:"lng"              description:"GPS longitude anchor"`
	ReleaseStage    string     `json:"releaseStage"    orm:"release_stage"    description:"Release stage: warmup, main, climax, closing"`
	OnlineAt        *time.Time `json:"onlineAt"        orm:"online_at"        description:"Scheduled online time"`
	VisibleWeekdays string     `json:"visibleWeekdays" orm:"visible_weekdays" description:"Optional visible weekdays, e.g. 1,3,5"`
	VisibleStart    string     `json:"visibleStart"    orm:"visible_start"    description:"Optional visible window start HH:MM"`
	VisibleEnd      string     `json:"visibleEnd"      orm:"visible_end"      description:"Optional visible window end HH:MM"`
	Status          string     `json:"status"          orm:"status"           description:"Cattle status: inactive, active; managed by activation flow"`
	CreatedAt       *time.Time `json:"createdAt"       orm:"created_at"       description:"Creation time"`
	UpdatedAt       *time.Time `json:"updatedAt"       orm:"updated_at"       description:"Update time"`
	DeletedAt       *time.Time `json:"deletedAt"       orm:"deleted_at"       description:"Soft-delete time, NULL means active"`
}
