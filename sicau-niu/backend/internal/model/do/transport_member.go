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
	g.Meta                  `orm:"table:plugin_sicau_niu_transport_member, do:true"`
	Id                      any        //
	TeamId                  any        //
	UserId                  any        //
	JoinRequestId           any        // Required joiner-scoped idempotency key; empty for creator membership
	Role                    any        // Member display role: creator, member
	JoinedAt                *time.Time //
	LeftAt                  *time.Time //
	CreatedAt               *time.Time //
	UpdatedAt               *time.Time //
	TotalContributionMeters any        //
	LastReportLat           any        //
	LastReportLng           any        //
	LastReportAt            *time.Time //
	JoinResponseJson        any        // Stable first successful team join response for idempotent replay
}
