// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"time"

	"github.com/gogf/gf/v2/frame/g"
)

// Sms is the golang structure of table plugin_linapro_uidentity_cas_sms for DAO operations like Where/Data.
type Sms struct {
	g.Meta    `orm:"table:plugin_linapro_uidentity_cas_sms, do:true"`
	Id        any        //
	Phone     any        //
	Type      any        //
	Content   any        //
	Status    any        //
	RespMsg   any        //
	CreatedAt *time.Time //
	UpdatedAt *time.Time //
	DeletedAt *time.Time //
	CreateBy  any        //
	UpdateBy  any        //
}
