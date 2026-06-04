// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

import (
	"time"
)

// Settlement is the golang structure for table settlement.
type Settlement struct {
	Id         int64      `json:"id"         orm:"id"          description:""`
	Title      string     `json:"title"      orm:"title"       description:"Archive title given by the operator"`
	Snapshot   string     `json:"snapshot"   orm:"snapshot"    description:"Frozen dashboard metrics serialized as JSON text"`
	OperatorId int64      `json:"operatorId" orm:"operator_id" description:"Host operator user ID who created the archive"`
	ArchivedAt *time.Time `json:"archivedAt" orm:"archived_at" description:"Archive time, set explicitly by the settlement service"`
	CreatedAt  *time.Time `json:"createdAt"  orm:"created_at"  description:"Creation time"`
	UpdatedAt  *time.Time `json:"updatedAt"  orm:"updated_at"  description:"Update time"`
	DeletedAt  *time.Time `json:"deletedAt"  orm:"deleted_at"  description:"Soft-delete time, NULL means active"`
}
