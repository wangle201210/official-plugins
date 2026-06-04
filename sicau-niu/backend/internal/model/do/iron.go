// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"time"

	"github.com/gogf/gf/v2/frame/g"
)

// Iron is the golang structure of table plugin_sicau_niu_iron for DAO operations like Where/Data.
type Iron struct {
	g.Meta    `orm:"table:plugin_sicau_niu_iron, do:true"`
	Id        any        //
	Code      any        // Iron-cow device identifier, unique among active rows
	Name      any        // Iron-cow display name
	LastLat   any        // Latest GPS latitude pulled by the bonus flow (C4)
	LastLng   any        // Latest GPS longitude pulled by the bonus flow (C4)
	LocatedAt *time.Time // Latest location pull time
	Remark    any        // Iron-cow remark
	CreatedAt *time.Time // Creation time
	UpdatedAt *time.Time // Update time
	DeletedAt *time.Time // Soft-delete time, NULL means active
}
