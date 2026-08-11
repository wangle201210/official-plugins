// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"time"

	"github.com/gogf/gf/v2/frame/g"
)

// TransportMember is the golang structure of table plugin_sicau_niu_transport_member for DAO operations like Where/Data.
type TransportMember struct {
	g.Meta          `orm:"table:plugin_sicau_niu_transport_member, do:true"`
	Id              any        //
	TeamId          any        //
	UserId          any        //
	JoinRequestId   any        //
	LeaveRequestId  any        //
	Role            any        // Member role: leader, member
	JoinedAt        *time.Time //
	LeftAt          *time.Time //
	LastHeartbeatAt *time.Time //
	CreatedAt       *time.Time //
	UpdatedAt       *time.Time //
}
