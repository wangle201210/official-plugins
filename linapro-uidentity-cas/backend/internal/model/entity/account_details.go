// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

import (
	"time"
)

// AccountDetails is the golang structure for table account_details.
type AccountDetails struct {
	AccountId int64      `json:"accountId" orm:"account_id" description:""`
	Birthday  *time.Time `json:"birthday"  orm:"birthday"   description:""`
	Email     string     `json:"email"     orm:"email"      description:""`
	Gender    int64      `json:"gender"    orm:"gender"     description:""`
	Qq        string     `json:"qq"        orm:"qq"         description:""`
	Wechat    string     `json:"wechat"    orm:"wechat"     description:""`
	Idcard    string     `json:"idcard"    orm:"idcard"     description:""`
	Avatar    string     `json:"avatar"    orm:"avatar"     description:""`
	Source    string     `json:"source"    orm:"source"     description:""`
	Nj        int64      `json:"nj"        orm:"nj"         description:""`
	Xymc      string     `json:"xymc"      orm:"xymc"       description:""`
	Xydm      string     `json:"xydm"      orm:"xydm"       description:""`
	Xq        string     `json:"xq"        orm:"xq"         description:""`
	Xz        int64      `json:"xz"        orm:"xz"         description:""`
	Yjbysj    int64      `json:"yjbysj"    orm:"yjbysj"     description:""`
	Zymc      string     `json:"zymc"      orm:"zymc"       description:""`
	Bjmc      string     `json:"bjmc"      orm:"bjmc"       description:""`
	Face      int64      `json:"face"      orm:"face"       description:""`
	CreatedAt *time.Time `json:"createdAt" orm:"created_at" description:""`
	UpdatedAt *time.Time `json:"updatedAt" orm:"updated_at" description:""`
	DeletedAt *time.Time `json:"deletedAt" orm:"deleted_at" description:""`
	CreateBy  int64      `json:"createBy"  orm:"create_by"  description:""`
	UpdateBy  int64      `json:"updateBy"  orm:"update_by"  description:""`
}
