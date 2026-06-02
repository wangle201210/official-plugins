// This file declares the old account update endpoint DTO.

package v1

import (
	"time"

	"github.com/gogf/gf/v2/frame/g"
)

// ResourceUpdateReq defines the old PUT /api/v1/account/{id} request.
type ResourceUpdateReq struct {
	g.Meta      `path:"/api/v1/account/{id}" method:"put" tags:"UIdentity CAS" summary:"Update account" dc:"Match the old uidentity/admin account update contract." permission:"uidentity:cas:write"`
	Id          int64     `json:"id" v:"required|min:1" dc:"Account ID from the legacy path parameter" eg:"1"`
	Number      string    `json:"number" dc:"Account number" eg:"A001"`
	Name        string    `json:"name" dc:"Account display name" eg:"Alice"`
	Phone       string    `json:"phone" dc:"Mobile phone number" eg:"13800000000"`
	EffectAt    time.Time `json:"effectAt" dc:"Account effective time" eg:"2026-06-02T00:00:00+08:00"`
	ExpireAt    time.Time `json:"expireAt" dc:"Account expiration time" eg:"2027-06-02T00:00:00+08:00"`
	GroupIds    []int64   `json:"groupIds" dc:"Account group IDs" eg:"[1,2]"`
	PassLevel   int64     `json:"passLevel" dc:"Password strength level" eg:"3"`
	ContainerId int64     `json:"containerId" dc:"LDAP container ID" eg:"1"`
	UnitId      int64     `json:"unitId" dc:"Unit ID" eg:"1"`
	Status      int64     `json:"status" dc:"Account status" eg:"1"`
}
