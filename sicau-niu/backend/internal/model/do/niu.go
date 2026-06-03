// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"time"

	"github.com/gogf/gf/v2/frame/g"
)

// Niu is the golang structure of table plugin_sicau_niu_niu for DAO operations like Where/Data.
type Niu struct {
	g.Meta          `orm:"table:plugin_sicau_niu_niu, do:true"`
	Id              any        //
	Code            any        // Cattle serial code, unique among active rows
	NiuType         any        // Cattle type: common, special
	SpecialSubtype  any        // Special subtype: college, contribution, alumni, spirit; empty for common
	Name            any        // Cattle name, used by special cattle
	CollegeId       any        // Linked college ID for college cattle, 0 means none
	Lat             any        // GPS latitude anchor
	Lng             any        // GPS longitude anchor
	ReleaseStage    any        // Release stage: warmup, main, climax, closing
	OnlineAt        *time.Time // Scheduled online time
	VisibleWeekdays any        // Optional visible weekdays, e.g. 1,3,5
	VisibleStart    any        // Optional visible window start HH:MM
	VisibleEnd      any        // Optional visible window end HH:MM
	Status          any        // Cattle status: inactive, active; managed by activation flow
	CreatedAt       *time.Time // Creation time
	UpdatedAt       *time.Time // Update time
	DeletedAt       *time.Time // Soft-delete time, NULL means active
}
