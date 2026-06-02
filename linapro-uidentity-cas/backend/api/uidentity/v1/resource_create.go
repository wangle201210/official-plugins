// This file declares the old account create endpoint DTO.

package v1

import (
	"time"

	"github.com/gogf/gf/v2/frame/g"
)

// ResourceCreateReq defines the old POST /api/v1/account request.
type ResourceCreateReq struct {
	g.Meta      `path:"/api/v1/account" method:"post" tags:"UIdentity CAS" summary:"Create account" dc:"Match the old uidentity/admin account create contract. The legacy create payload uses groupId as the JSON field for UnitId and groupIds for account-group membership." permission:"uidentity:cas:write"`
	Number      string    `json:"number" dc:"Account number" eg:"A001"`
	Name        string    `json:"name" dc:"Account display name" eg:"Alice"`
	Phone       string    `json:"phone" dc:"Mobile phone number" eg:"13800000000"`
	EffectAt    time.Time `json:"effectAt" dc:"Account effective time" eg:"2026-06-02T00:00:00+08:00"`
	ExpireAt    time.Time `json:"expireAt" dc:"Account expiration time" eg:"2027-06-02T00:00:00+08:00"`
	UnitId      int64     `json:"groupId" dc:"Legacy create payload field for unit ID" eg:"1"`
	PassLevel   int64     `json:"passLevel" dc:"Password strength level" eg:"3"`
	ContainerId int64     `json:"containerId" dc:"LDAP container ID" eg:"1"`
	GroupIDs    []int64   `json:"groupIds" dc:"Account group IDs" eg:"[1,2]"`
	Status      int64     `json:"status" dc:"Account status" eg:"1"`
}
