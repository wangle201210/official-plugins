// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

// AccountUnit is the golang structure for table account_unit.
type AccountUnit struct {
	Id        int64 `json:"id"        orm:"id"         description:""`
	AccountId int64 `json:"accountId" orm:"account_id" description:""`
	UnitId    int64 `json:"unitId"    orm:"unit_id"    description:""`
}
