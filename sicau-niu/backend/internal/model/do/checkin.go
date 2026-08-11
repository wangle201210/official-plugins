// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"time"

	"github.com/gogf/gf/v2/frame/g"
)

// Checkin is the golang structure of table plugin_sicau_niu_checkin for DAO operations like Where/Data.
type Checkin struct {
	g.Meta        `orm:"table:plugin_sicau_niu_checkin, do:true"`
	Id            any        //
	UserId        any        //
	CheckinDate   any        // Check-in date YYYY-MM-DD
	Amount        any        // Granted grass amount (20-50)
	CreatedAt     *time.Time //
	UpdatedAt     *time.Time //
	DeletedAt     *time.Time //
	RequestId     any        // Player-scoped idempotency key
	ResultBalance any        // Stable account balance returned by the first successful request
}
