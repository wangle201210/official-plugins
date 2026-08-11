// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

import (
	"time"
)

// TransportMember is the golang structure for table transport_member.
type TransportMember struct {
	Id              int64      `json:"id"              orm:"id"                description:""`
	TeamId          int64      `json:"teamId"          orm:"team_id"           description:""`
	UserId          int64      `json:"userId"          orm:"user_id"           description:""`
	JoinRequestId   string     `json:"joinRequestId"   orm:"join_request_id"   description:""`
	LeaveRequestId  string     `json:"leaveRequestId"  orm:"leave_request_id"  description:""`
	Role            string     `json:"role"            orm:"role"              description:"Member role: leader, member"`
	JoinedAt        *time.Time `json:"joinedAt"        orm:"joined_at"         description:""`
	LeftAt          *time.Time `json:"leftAt"          orm:"left_at"           description:""`
	LastHeartbeatAt *time.Time `json:"lastHeartbeatAt" orm:"last_heartbeat_at" description:""`
	CreatedAt       *time.Time `json:"createdAt"       orm:"created_at"        description:""`
	UpdatedAt       *time.Time `json:"updatedAt"       orm:"updated_at"        description:""`
}
