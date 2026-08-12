// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"time"

	"github.com/gogf/gf/v2/frame/g"
)

// TransportReport is the golang structure of table plugin_sicau_niu_transport_report for DAO operations like Where/Data.
type TransportReport struct {
	g.Meta             `orm:"table:plugin_sicau_niu_transport_report, do:true"`
	Id                 any        //
	ActivityKey        any        //
	TeamId             any        //
	MemberId           any        //
	UserId             any        //
	RequestId          any        //
	ActivityDate       any        // Beijing natural-day key derived from accepted_at
	StartLat           any        // Previous successful report latitude; NULL for the first report in a membership
	StartLng           any        //
	EndLat             any        //
	EndLng             any        //
	SampledAt          *time.Time //
	AcceptedAt         *time.Time //
	ContributionMeters any        //
	UserTotalMeters    any        //
	TeamTotalMeters    any        //
	DailyReportCount   any        //
	CreatedAt          *time.Time //
}
