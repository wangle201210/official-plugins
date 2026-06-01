// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

import (
	"time"
)

// AccountDetail is the golang structure for table account_detail.
type AccountDetail struct {
	AccountId    int64      `json:"accountId"    orm:"account_id"    description:""`
	TenantId     int        `json:"tenantId"     orm:"tenant_id"     description:""`
	Birthday     string     `json:"birthday"     orm:"birthday"      description:"Date-only birthday in YYYY-MM-DD format"`
	Email        string     `json:"email"        orm:"email"         description:""`
	Gender       int        `json:"gender"       orm:"gender"        description:""`
	Qq           string     `json:"qq"           orm:"qq"            description:""`
	Wechat       string     `json:"wechat"       orm:"wechat"        description:""`
	Idcard       string     `json:"idcard"       orm:"idcard"        description:""`
	Avatar       string     `json:"avatar"       orm:"avatar"        description:""`
	Source       string     `json:"source"       orm:"source"        description:""`
	Grade        string     `json:"grade"        orm:"grade"         description:""`
	College      string     `json:"college"      orm:"college"       description:""`
	CollegeCode  string     `json:"collegeCode"  orm:"college_code"  description:""`
	Campus       string     `json:"campus"       orm:"campus"        description:""`
	SchoolSystem string     `json:"schoolSystem" orm:"school_system" description:""`
	GraduatedAt  string     `json:"graduatedAt"  orm:"graduated_at"  description:""`
	Major        string     `json:"major"        orm:"major"         description:""`
	ClassName    string     `json:"className"    orm:"class_name"    description:""`
	Face         string     `json:"face"         orm:"face"          description:""`
	CreatedBy    int64      `json:"createdBy"    orm:"created_by"    description:""`
	UpdatedBy    int64      `json:"updatedBy"    orm:"updated_by"    description:""`
	CreatedAt    *time.Time `json:"createdAt"    orm:"created_at"    description:""`
	UpdatedAt    *time.Time `json:"updatedAt"    orm:"updated_at"    description:""`
}
