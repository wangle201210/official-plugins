// admin_iron_reporting_cycle_update.go defines the operator action that updates
// one locator's external reporting cycle to the fixed 10-second activity value.

package v1

import "github.com/gogf/gf/v2/frame/g"

// UpdateIronReportingCycleReq is the request for setting one locator's reporting
// cycle to 10 seconds before the local iron-cow registration is created.
type UpdateIronReportingCycleReq struct {
	g.Meta `path:"/plugins/sicau-niu/admin/iron/reporting-cycle" method:"put" tags:"Sicau Niu Admin" summary:"设置铁牛 10 秒上报频率" dc:"Set one IOT locator's reporting cycle to the fixed 10-second activity value. This updates only the external locator configuration and does not create a local iron-cow record. Protected by host unified permission check." permission:"sicau-niu:iron:create"`
	Code   string `json:"code" v:"required|length:1,64" dc:"IOT locator device identifier entered in the new iron-cow form" eg:"50275156712"`
}

// UpdateIronReportingCycleRes is the response for a successful reporting-cycle
// update. It is empty because success is conveyed by the absence of an error.
type UpdateIronReportingCycleRes struct{}
