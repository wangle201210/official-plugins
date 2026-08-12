// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

import (
	"time"
)

// TransportReport is the golang structure for table transport_report.
type TransportReport struct {
	Id                 int64      `json:"id"                 orm:"id"                  description:""`
	ActivityKey        string     `json:"activityKey"        orm:"activity_key"        description:""`
	TeamId             int64      `json:"teamId"             orm:"team_id"             description:""`
	MemberId           int64      `json:"memberId"           orm:"member_id"           description:""`
	UserId             int64      `json:"userId"             orm:"user_id"             description:""`
	RequestId          string     `json:"requestId"          orm:"request_id"          description:""`
	ActivityDate       string     `json:"activityDate"       orm:"activity_date"       description:"Beijing natural-day key derived from accepted_at"`
	StartLat           float64    `json:"startLat"           orm:"start_lat"           description:"Previous successful report latitude; NULL for the first report in a membership"`
	StartLng           float64    `json:"startLng"           orm:"start_lng"           description:""`
	EndLat             float64    `json:"endLat"             orm:"end_lat"             description:""`
	EndLng             float64    `json:"endLng"             orm:"end_lng"             description:""`
	SampledAt          *time.Time `json:"sampledAt"          orm:"sampled_at"          description:""`
	AcceptedAt         *time.Time `json:"acceptedAt"         orm:"accepted_at"         description:""`
	ContributionMeters int64      `json:"contributionMeters" orm:"contribution_meters" description:""`
	UserTotalMeters    int64      `json:"userTotalMeters"    orm:"user_total_meters"   description:""`
	TeamTotalMeters    int64      `json:"teamTotalMeters"    orm:"team_total_meters"   description:""`
	DailyReportCount   int        `json:"dailyReportCount"   orm:"daily_report_count"  description:""`
	CreatedAt          *time.Time `json:"createdAt"          orm:"created_at"          description:""`
}
