// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

import (
	"time"
)

// TransportSession is the golang structure for table transport_session.
type TransportSession struct {
	Id              int64      `json:"id"              orm:"id"                 description:""`
	TeamId          int64      `json:"teamId"          orm:"team_id"            description:""`
	IronId          int64      `json:"ironId"          orm:"iron_id"            description:""`
	StartedByUserId int64      `json:"startedByUserId" orm:"started_by_user_id" description:""`
	StartRequestId  string     `json:"startRequestId"  orm:"start_request_id"   description:""`
	EndedByUserId   int64      `json:"endedByUserId"   orm:"ended_by_user_id"   description:""`
	EndRequestId    string     `json:"endRequestId"    orm:"end_request_id"     description:""`
	Status          string     `json:"status"          orm:"status"             description:"Session status: active, idle_timeout, ended"`
	StartedAt       *time.Time `json:"startedAt"       orm:"started_at"         description:""`
	LastActiveAt    *time.Time `json:"lastActiveAt"    orm:"last_active_at"     description:""`
	EndedAt         *time.Time `json:"endedAt"         orm:"ended_at"           description:""`
	MovedMeters     float64    `json:"movedMeters"     orm:"moved_meters"       description:""`
	LastLat         float64    `json:"lastLat"         orm:"last_lat"           description:""`
	LastLng         float64    `json:"lastLng"         orm:"last_lng"           description:""`
	CreatedAt       *time.Time `json:"createdAt"       orm:"created_at"         description:""`
	UpdatedAt       *time.Time `json:"updatedAt"       orm:"updated_at"         description:""`
}
