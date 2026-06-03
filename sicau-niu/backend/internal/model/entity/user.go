// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

import (
	"time"
)

// User is the golang structure for table user.
type User struct {
	Id                int64      `json:"id"                orm:"id"                 description:"Primary key ID"`
	Openid            string     `json:"openid"            orm:"openid"             description:"WeChat openid, unique per active player"`
	Phone             string     `json:"phone"             orm:"phone"              description:"Bound phone number, unique per active player (one-phone-one-account)"`
	Nickname          string     `json:"nickname"          orm:"nickname"           description:"Player nickname"`
	Avatar            string     `json:"avatar"            orm:"avatar"             description:"Player avatar URL"`
	IdentityType      string     `json:"identityType"      orm:"identity_type"      description:"Identity tag: student / alumni / friend"`
	CollegeId         int64      `json:"collegeId"         orm:"college_id"         description:"Selected college ID, 0 means none"`
	Grade             int        `json:"grade"             orm:"grade"              description:"Grade number filled by student, 0 means unset"`
	GraduationYear    int        `json:"graduationYear"    orm:"graduation_year"    description:"Graduation year filled by alumni, 0 means unset"`
	DeviceFingerprint string     `json:"deviceFingerprint" orm:"device_fingerprint" description:"Lightweight device fingerprint for risk control"`
	CreatedAt         *time.Time `json:"createdAt"         orm:"created_at"         description:"Creation time"`
	UpdatedAt         *time.Time `json:"updatedAt"         orm:"updated_at"         description:"Update time"`
	DeletedAt         *time.Time `json:"deletedAt"         orm:"deleted_at"         description:"Soft-delete time, NULL means active"`
}
