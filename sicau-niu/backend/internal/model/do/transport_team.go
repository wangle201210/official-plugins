// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"time"

	"github.com/gogf/gf/v2/frame/g"
)

// TransportTeam is the golang structure of table plugin_sicau_niu_transport_team for DAO operations like Where/Data.
type TransportTeam struct {
	g.Meta                  `orm:"table:plugin_sicau_niu_transport_team, do:true"`
	Id                      any        //
	Name                    any        //
	LeaderUserId            any        // Team creator player ID; creator has no lifecycle authority
	CreateRequestId         any        //
	Status                  any        // Team status: effective, invalid
	Visible                 any        //
	CreatedAt               *time.Time //
	UpdatedAt               *time.Time //
	DeletedAt               *time.Time //
	MemberCount             any        //
	TotalContributionMeters any        //
	LastActiveAt            *time.Time // Latest successful create, join or contribution server commit time
	InvalidatedAt           *time.Time //
	InvalidReason           any        //
}
