// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"time"

	"github.com/gogf/gf/v2/frame/g"
)

// Feeding is the golang structure of table plugin_sicau_niu_feeding for DAO operations like Where/Data.
type Feeding struct {
	g.Meta           `orm:"table:plugin_sicau_niu_feeding, do:true"`
	Id               any        //
	UserId           any        //
	NiuId            any        //
	BaseAmount       any        // Original grass amount fed
	CoefficientBasis any        // Bonus coefficient in basis of 100: 100=x1.0, 150=x1.5
	EffectAmount     any        // Actual feeding effect = base_amount * coefficient_basis / 100
	IsIronBonus      any        // Whether an iron-cow proximity bonus applied: 1 yes, 0 no
	FedAt            *time.Time //
	CreatedAt        *time.Time //
	UpdatedAt        *time.Time //
	DeletedAt        *time.Time //
	RequestId        any        // Client idempotency key deduplicating network retries; empty when not provided
}
