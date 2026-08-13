// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

import (
	"time"
)

// TransportMember is the golang structure for table transport_member.
type TransportMember struct {
	Id                      int64      `json:"id"                      orm:"id"                        description:""`
	TeamId                  int64      `json:"teamId"                  orm:"team_id"                   description:""`
	UserId                  int64      `json:"userId"                  orm:"user_id"                   description:""`
	JoinRequestId           string     `json:"joinRequestId"           orm:"join_request_id"           description:"Required joiner-scoped idempotency key; empty for creator membership"`
	Role                    string     `json:"role"                    orm:"role"                      description:"Member display role: creator, member"`
	JoinedAt                *time.Time `json:"joinedAt"                orm:"joined_at"                 description:""`
	LeftAt                  *time.Time `json:"leftAt"                  orm:"left_at"                   description:""`
	CreatedAt               *time.Time `json:"createdAt"               orm:"created_at"                description:""`
	UpdatedAt               *time.Time `json:"updatedAt"               orm:"updated_at"                description:""`
	TotalContributionMeters int64      `json:"totalContributionMeters" orm:"total_contribution_meters" description:""`
	LastReportLat           float64    `json:"lastReportLat"           orm:"last_report_lat"           description:""`
	LastReportLng           float64    `json:"lastReportLng"           orm:"last_report_lng"           description:""`
	LastReportAt            *time.Time `json:"lastReportAt"            orm:"last_report_at"            description:""`
	JoinResponseJson        string     `json:"joinResponseJson"        orm:"join_response_json"        description:"Stable first successful team join response for idempotent replay"`
}
