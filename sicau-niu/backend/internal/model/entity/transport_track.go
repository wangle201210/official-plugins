// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

import (
	"time"
)

// TransportTrack is the golang structure for table transport_track.
type TransportTrack struct {
	Id             int64      `json:"id"             orm:"id"              description:""`
	SessionId      int64      `json:"sessionId"      orm:"session_id"      description:""`
	UserId         int64      `json:"userId"         orm:"user_id"         description:""`
	RequestId      string     `json:"requestId"      orm:"request_id"      description:""`
	Lat            float64    `json:"lat"            orm:"lat"             description:""`
	Lng            float64    `json:"lng"            orm:"lng"             description:""`
	DistanceMeters float64    `json:"distanceMeters" orm:"distance_meters" description:""`
	RecordedAt     *time.Time `json:"recordedAt"     orm:"recorded_at"     description:""`
	CreatedAt      *time.Time `json:"createdAt"      orm:"created_at"      description:""`
}
