// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"time"

	"github.com/gogf/gf/v2/frame/g"
)

// Card is the golang structure of table plugin_sicau_niu_card for DAO operations like Where/Data.
type Card struct {
	g.Meta    `orm:"table:plugin_sicau_niu_card, do:true"`
	Id        any        //
	NiuId     any        // Owning cattle ID; one card per cattle
	Category  any        // Card category: person, event, research, college, spirit
	Title     any        // Card title
	Content   any        // Card content text
	ImagePath any        // Card image storage path uploaded via host file management
	CreatedAt *time.Time // Creation time
	UpdatedAt *time.Time // Update time
	DeletedAt *time.Time // Soft-delete time, NULL means active
}
