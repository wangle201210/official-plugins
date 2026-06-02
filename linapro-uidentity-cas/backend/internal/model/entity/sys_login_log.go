// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

import (
	"time"
)

// SysLoginLog is the golang structure for table sys_login_log.
type SysLoginLog struct {
	Id            int64      `json:"id"            orm:"id"             description:""`
	Username      string     `json:"username"      orm:"username"       description:""`
	Status        string     `json:"status"        orm:"status"         description:""`
	Ipaddr        string     `json:"ipaddr"        orm:"ipaddr"         description:""`
	LoginLocation string     `json:"loginLocation" orm:"login_location" description:""`
	Browser       string     `json:"browser"       orm:"browser"        description:""`
	Os            string     `json:"os"            orm:"os"             description:""`
	Platform      string     `json:"platform"      orm:"platform"       description:""`
	LoginTime     *time.Time `json:"loginTime"     orm:"login_time"     description:""`
	Remark        string     `json:"remark"        orm:"remark"         description:""`
	Msg           string     `json:"msg"           orm:"msg"            description:""`
	CreatedAt     *time.Time `json:"createdAt"     orm:"created_at"     description:""`
	UpdatedAt     *time.Time `json:"updatedAt"     orm:"updated_at"     description:""`
	CreateBy      int64      `json:"createBy"      orm:"create_by"      description:""`
	UpdateBy      int64      `json:"updateBy"      orm:"update_by"      description:""`
}
