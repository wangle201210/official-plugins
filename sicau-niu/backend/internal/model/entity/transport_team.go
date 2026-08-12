// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

import (
	"time"
)

// TransportTeam is the golang structure for table transport_team.
type TransportTeam struct {
	Id                      int64      `json:"id"                      orm:"id"                        description:""`
	Name                    string     `json:"name"                    orm:"name"                      description:""`
	LeaderUserId            int64      `json:"leaderUserId"            orm:"leader_user_id"            description:"Team creator player ID; creator has no lifecycle authority"`
	CreateRequestId         string     `json:"createRequestId"         orm:"create_request_id"         description:""`
	Status                  string     `json:"status"                  orm:"status"                    description:"Team status: effective, invalid"`
	Visible                 int        `json:"visible"                 orm:"visible"                   description:""`
	CreatedAt               *time.Time `json:"createdAt"               orm:"created_at"                description:""`
	UpdatedAt               *time.Time `json:"updatedAt"               orm:"updated_at"                description:""`
	DeletedAt               *time.Time `json:"deletedAt"               orm:"deleted_at"                description:""`
	MemberCount             int        `json:"memberCount"             orm:"member_count"              description:""`
	TotalContributionMeters int64      `json:"totalContributionMeters" orm:"total_contribution_meters" description:""`
	LastActiveAt            *time.Time `json:"lastActiveAt"            orm:"last_active_at"            description:"Latest successful create, join or contribution server commit time"`
	InvalidatedAt           *time.Time `json:"invalidatedAt"           orm:"invalidated_at"            description:""`
	InvalidReason           string     `json:"invalidReason"           orm:"invalid_reason"            description:""`
}
