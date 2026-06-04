// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"time"

	"github.com/gogf/gf/v2/frame/g"
)

// HonorDef is the golang structure of table plugin_sicau_niu_honor_def for DAO operations like Where/Data.
type HonorDef struct {
	g.Meta     `orm:"table:plugin_sicau_niu_honor_def, do:true"`
	Id         any        //
	HonorType  any        // Honor type: badge, avatar_frame, certificate
	Code       any        // Honor unique code among active rows
	Name       any        //
	UnlockType any        // Unlock rule: participation, feed_count, activation_count, category_complete, full_complete
	Threshold  any        // Threshold for count-based unlock rules
	Category   any        // Card category for category_complete unlock rule
	ImagePath  any        // Honor image/template storage path
	Sort       any        //
	CreatedAt  *time.Time //
	UpdatedAt  *time.Time //
	DeletedAt  *time.Time // Soft-delete time, NULL means active
}
