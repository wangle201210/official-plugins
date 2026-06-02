// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

import (
	"time"
)

// SysOperaLog is the golang structure for table sys_opera_log.
type SysOperaLog struct {
	Id            int64      `json:"id"            orm:"id"             description:""`
	Title         string     `json:"title"         orm:"title"          description:""`
	BusinessType  string     `json:"businessType"  orm:"business_type"  description:""`
	BusinessTypes string     `json:"businessTypes" orm:"business_types" description:""`
	Method        string     `json:"method"        orm:"method"         description:""`
	RequestMethod string     `json:"requestMethod" orm:"request_method" description:""`
	OperatorType  string     `json:"operatorType"  orm:"operator_type"  description:""`
	OperName      string     `json:"operName"      orm:"oper_name"      description:""`
	DeptName      string     `json:"deptName"      orm:"dept_name"      description:""`
	OperUrl       string     `json:"operUrl"       orm:"oper_url"       description:""`
	OperIp        string     `json:"operIp"        orm:"oper_ip"        description:""`
	OperLocation  string     `json:"operLocation"  orm:"oper_location"  description:""`
	OperParam     string     `json:"operParam"     orm:"oper_param"     description:""`
	Status        string     `json:"status"        orm:"status"         description:""`
	OperTime      *time.Time `json:"operTime"      orm:"oper_time"      description:""`
	JsonResult    string     `json:"jsonResult"    orm:"json_result"    description:""`
	Remark        string     `json:"remark"        orm:"remark"         description:""`
	LatencyTime   string     `json:"latencyTime"   orm:"latency_time"   description:""`
	UserAgent     string     `json:"userAgent"     orm:"user_agent"     description:""`
	CreatedAt     *time.Time `json:"createdAt"     orm:"created_at"     description:""`
	UpdatedAt     *time.Time `json:"updatedAt"     orm:"updated_at"     description:""`
	CreateBy      int64      `json:"createBy"      orm:"create_by"      description:""`
	UpdateBy      int64      `json:"updateBy"      orm:"update_by"      description:""`
}
