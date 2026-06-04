// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

import (
	"time"
)

// Feeding is the golang structure for table feeding.
type Feeding struct {
	Id               int64      `json:"id"               orm:"id"                description:""`
	UserId           int64      `json:"userId"           orm:"user_id"           description:""`
	NiuId            int64      `json:"niuId"            orm:"niu_id"            description:""`
	BaseAmount       int        `json:"baseAmount"       orm:"base_amount"       description:"Original grass amount fed"`
	CoefficientBasis int        `json:"coefficientBasis" orm:"coefficient_basis" description:"Bonus coefficient in basis of 100: 100=x1.0, 150=x1.5"`
	EffectAmount     int        `json:"effectAmount"     orm:"effect_amount"     description:"Actual feeding effect = base_amount * coefficient_basis / 100"`
	IsIronBonus      int        `json:"isIronBonus"      orm:"is_iron_bonus"     description:"Whether an iron-cow proximity bonus applied: 1 yes, 0 no"`
	FedAt            *time.Time `json:"fedAt"            orm:"fed_at"            description:""`
	CreatedAt        *time.Time `json:"createdAt"        orm:"created_at"        description:""`
	UpdatedAt        *time.Time `json:"updatedAt"        orm:"updated_at"        description:""`
	DeletedAt        *time.Time `json:"deletedAt"        orm:"deleted_at"        description:""`
}
