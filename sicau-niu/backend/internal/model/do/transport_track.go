// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"time"

	"github.com/gogf/gf/v2/frame/g"
)

// TransportTrack is the golang structure of table plugin_sicau_niu_transport_track for DAO operations like Where/Data.
type TransportTrack struct {
	g.Meta         `orm:"table:plugin_sicau_niu_transport_track, do:true"`
	Id             any        //
	SessionId      any        //
	UserId         any        //
	RequestId      any        //
	Lat            any        //
	Lng            any        //
	DistanceMeters any        //
	RecordedAt     *time.Time //
	CreatedAt      *time.Time //
}
