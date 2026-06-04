// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"time"

	"github.com/gogf/gf/v2/frame/g"
)

// UserHonor is the golang structure of table plugin_sicau_niu_user_honor for DAO operations like Where/Data.
type UserHonor struct {
	g.Meta     `orm:"table:plugin_sicau_niu_user_honor, do:true"`
	Id         any        //
	UserId     any        // Player ID
	HonorId    any        // Granted honor definition ID
	UnlockedAt *time.Time // Grant time
	CreatedAt  *time.Time //
	UpdatedAt  *time.Time //
	DeletedAt  *time.Time //
}
