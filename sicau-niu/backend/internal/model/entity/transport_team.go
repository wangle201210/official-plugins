// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

import (
	"time"
)

// TransportTeam is the golang structure for table transport_team.
type TransportTeam struct {
	Id              int64      `json:"id"              orm:"id"                description:""`
	Code            string     `json:"code"            orm:"code"              description:""`
	Name            string     `json:"name"            orm:"name"              description:""`
	CampusId        string     `json:"campusId"        orm:"campus_id"         description:""`
	LeaderUserId    int64      `json:"leaderUserId"    orm:"leader_user_id"    description:""`
	CreateRequestId string     `json:"createRequestId" orm:"create_request_id" description:""`
	IronId          int64      `json:"ironId"          orm:"iron_id"           description:""`
	Status          string     `json:"status"          orm:"status"            description:"Team status: forming, active, ended"`
	MinMembers      int        `json:"minMembers"      orm:"min_members"       description:""`
	MaxMembers      int        `json:"maxMembers"      orm:"max_members"       description:""`
	Visible         int        `json:"visible"         orm:"visible"           description:""`
	CreatedAt       *time.Time `json:"createdAt"       orm:"created_at"        description:""`
	UpdatedAt       *time.Time `json:"updatedAt"       orm:"updated_at"        description:""`
	DeletedAt       *time.Time `json:"deletedAt"       orm:"deleted_at"        description:""`
}
