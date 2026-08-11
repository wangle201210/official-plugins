// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"time"

	"github.com/gogf/gf/v2/frame/g"
)

// TransportSession is the golang structure of table plugin_sicau_niu_transport_session for DAO operations like Where/Data.
type TransportSession struct {
	g.Meta          `orm:"table:plugin_sicau_niu_transport_session, do:true"`
	Id              any        //
	TeamId          any        //
	IronId          any        //
	StartedByUserId any        //
	StartRequestId  any        //
	EndedByUserId   any        //
	EndRequestId    any        //
	Status          any        // Session status: active, idle_timeout, ended
	StartedAt       *time.Time //
	LastActiveAt    *time.Time //
	EndedAt         *time.Time //
	MovedMeters     any        //
	LastLat         any        //
	LastLng         any        //
	CreatedAt       *time.Time //
	UpdatedAt       *time.Time //
}
