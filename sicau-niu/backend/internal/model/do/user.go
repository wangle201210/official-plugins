// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"time"

	"github.com/gogf/gf/v2/frame/g"
)

// User is the golang structure of table plugin_sicau_niu_user for DAO operations like Where/Data.
type User struct {
	g.Meta            `orm:"table:plugin_sicau_niu_user, do:true"`
	Id                any        // Primary key ID
	Openid            any        // WeChat openid, unique per active player
	Phone             any        // Bound phone number, unique per active player (one-phone-one-account)
	Nickname          any        // Player nickname
	Avatar            any        // Player avatar URL
	IdentityType      any        // Identity tag: student / alumni / friend
	CollegeId         any        // Selected college ID, 0 means none
	Grade             any        // Grade number filled by student, 0 means unset
	GraduationYear    any        // Graduation year filled by alumni, 0 means unset
	DeviceFingerprint any        // Lightweight device fingerprint for risk control
	CreatedAt         *time.Time // Creation time
	UpdatedAt         *time.Time // Update time
	DeletedAt         *time.Time // Soft-delete time, NULL means active
}
