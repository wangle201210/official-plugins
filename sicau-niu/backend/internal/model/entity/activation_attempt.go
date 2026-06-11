// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

import (
	"time"
)

// ActivationAttempt is the golang structure for table activation_attempt.
type ActivationAttempt struct {
	Id           int64      `json:"id"           orm:"id"             description:""`
	UserId       int64      `json:"userId"       orm:"user_id"        description:"Player ID that submitted the photo check-in"`
	NiuId        int64      `json:"niuId"        orm:"niu_id"         description:"Activated cattle ID when result is success, otherwise 0"`
	NearestNiuId int64      `json:"nearestNiuId" orm:"nearest_niu_id" description:"Nearest visible inactive cattle candidate ID when available"`
	Result       string     `json:"result"       orm:"result"         description:"Attempt result: success, no_nearby, out_of_range"`
	Lat          float64    `json:"lat"          orm:"lat"            description:"Player reported GPS latitude"`
	Lng          float64    `json:"lng"          orm:"lng"            description:"Player reported GPS longitude"`
	DistanceM    float64    `json:"distanceM"    orm:"distance_m"     description:"Distance in meters to nearest candidate, 0 when unavailable"`
	ThresholdM   float64    `json:"thresholdM"   orm:"threshold_m"    description:"LBS activation threshold in meters used for the attempt"`
	PhotoPath    string     `json:"photoPath"    orm:"photo_path"     description:"Uploaded photo storage path, evidence only"`
	AttemptedAt  *time.Time `json:"attemptedAt"  orm:"attempted_at"   description:"Attempt time"`
	CreatedAt    *time.Time `json:"createdAt"    orm:"created_at"     description:"Creation time"`
	UpdatedAt    *time.Time `json:"updatedAt"    orm:"updated_at"     description:"Update time"`
	DeletedAt    *time.Time `json:"deletedAt"    orm:"deleted_at"     description:"Soft-delete time, NULL means active"`
}
