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
	g.Meta          `orm:"table:plugin_sicau_niu_transport_team, do:true"`
	Id              any        //
	Code            any        //
	Name            any        //
	CampusId        any        //
	LeaderUserId    any        //
	CreateRequestId any        //
	IronId          any        //
	Status          any        // Team status: forming, active, ended
	MinMembers      any        //
	MaxMembers      any        //
	Visible         any        //
	CreatedAt       *time.Time //
	UpdatedAt       *time.Time //
	DeletedAt       *time.Time //
}
