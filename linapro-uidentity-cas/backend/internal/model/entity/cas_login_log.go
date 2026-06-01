// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

import (
	"time"
)

// CasLoginLog is the golang structure for table cas_login_log.
type CasLoginLog struct {
	Id              int64      `json:"id"              orm:"id"                description:""`
	TenantId        int        `json:"tenantId"        orm:"tenant_id"         description:""`
	AccountId       int64      `json:"accountId"       orm:"account_id"        description:""`
	ChoiceAccountId int64      `json:"choiceAccountId" orm:"choice_account_id" description:""`
	AppId           int64      `json:"appId"           orm:"app_id"            description:""`
	Ipaddr          string     `json:"ipaddr"          orm:"ipaddr"            description:""`
	LoginLocation   string     `json:"loginLocation"   orm:"login_location"    description:""`
	Browser         string     `json:"browser"         orm:"browser"           description:""`
	Os              string     `json:"os"              orm:"os"                description:""`
	Platform        string     `json:"platform"        orm:"platform"          description:""`
	LoginTime       *time.Time `json:"loginTime"       orm:"login_time"        description:""`
	Remark          string     `json:"remark"          orm:"remark"            description:""`
	Msg             string     `json:"msg"             orm:"msg"               description:""`
	LoginType       string     `json:"loginType"       orm:"login_type"        description:""`
	CreatedBy       int64      `json:"createdBy"       orm:"created_by"        description:""`
	UpdatedBy       int64      `json:"updatedBy"       orm:"updated_by"        description:""`
	CreatedAt       *time.Time `json:"createdAt"       orm:"created_at"        description:""`
	UpdatedAt       *time.Time `json:"updatedAt"       orm:"updated_at"        description:""`
	DeletedAt       *time.Time `json:"deletedAt"       orm:"deleted_at"        description:""`
}
