// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"time"

	"github.com/gogf/gf/v2/frame/g"
)

// ActivationPhoto is the golang structure of table plugin_sicau_niu_activation_photo for DAO operations like Where/Data.
type ActivationPhoto struct {
	g.Meta       `orm:"table:plugin_sicau_niu_activation_photo, do:true"`
	Id           any        //
	Token        any        // Opaque unguessable player-facing photo identifier
	UserId       any        // Owning player ID
	RequestId    any        // Player-scoped idempotency key
	ActivityDate any        // Beijing natural-day key
	DailySlot    any        // Bounded successful upload slot 1..10 within the activity date
	ObjectPath   any        // Plugin-private logical object storage path
	OriginalName any        //
	ContentType  any        //
	SizeBytes    any        //
	UsedAt       *time.Time // Successful activation consumption time; NULL means unused
	CreatedAt    *time.Time //
	UpdatedAt    *time.Time //
}
