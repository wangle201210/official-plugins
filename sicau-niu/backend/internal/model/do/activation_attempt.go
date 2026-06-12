// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"time"

	"github.com/gogf/gf/v2/frame/g"
)

// ActivationAttempt is the golang structure of table plugin_sicau_niu_activation_attempt for DAO operations like Where/Data.
type ActivationAttempt struct {
	g.Meta       `orm:"table:plugin_sicau_niu_activation_attempt, do:true"`
	Id           any        //
	UserId       any        // Player ID that submitted the photo check-in
	NiuId        any        // Activated cattle ID when result is success, otherwise 0
	NearestNiuId any        // Nearest visible inactive cattle candidate ID when available
	Result       any        // Attempt result: success, no_nearby, out_of_range, speed_anomaly
	Lat          any        // Player reported GPS latitude
	Lng          any        // Player reported GPS longitude
	DistanceM    any        // Distance in meters to nearest candidate, 0 when unavailable
	ThresholdM   any        // LBS activation threshold in meters used for the attempt
	PhotoPath    any        // Uploaded photo storage path, evidence only
	AttemptedAt  *time.Time // Attempt time
	CreatedAt    *time.Time // Creation time
	UpdatedAt    *time.Time // Update time
	DeletedAt    *time.Time // Soft-delete time, NULL means active
}
