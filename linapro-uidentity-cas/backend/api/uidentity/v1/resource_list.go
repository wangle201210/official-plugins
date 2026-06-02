// This file declares the old account list endpoint DTO.

package v1

import "github.com/gogf/gf/v2/frame/g"

// ResourceListReq defines the old GET /api/v1/account request.
type ResourceListReq struct {
	g.Meta           `path:"/api/v1/account" method:"get" tags:"UIdentity CAS" summary:"List accounts" dc:"Match the old uidentity/admin account list contract. Query parameters use the legacy pageIndex/pageSize, filter, array and fieldOrder names." permission:"uidentity:cas:read"`
	PageNum          int     `json:"pageIndex" d:"1" v:"min:1" dc:"Legacy page number starting from 1" eg:"1"`
	PageSize         int     `json:"pageSize" d:"10" v:"min:1|max:100" dc:"Legacy page size with plugin safety cap" eg:"10"`
	Number           string  `json:"number" dc:"Exact account number filter" eg:"A001"`
	Name             string  `json:"name" dc:"Fuzzy account display name filter" eg:"Alice"`
	Phone            string  `json:"phone" dc:"Fuzzy mobile phone filter" eg:"13800000000"`
	GroupIds         []int64 `json:"groupIds" dc:"Legacy repeated groupIds query filter" eg:"[1,2]"`
	PassLevels       []int64 `json:"passLevels" dc:"Legacy repeated passLevels query filter" eg:"[3,4]"`
	ContainerId      int64   `json:"containerId" dc:"LDAP container ID filter" eg:"1"`
	UnitId           int64   `json:"unitId" dc:"Unit ID filter" eg:"1"`
	Status           *int    `json:"status" dc:"Account status filter" eg:"1"`
	IdOrder          string  `json:"idOrder" dc:"Legacy id sort direction" eg:"desc"`
	NumberOrder      string  `json:"numberOrder" dc:"Legacy number sort direction" eg:"asc"`
	NameOrder        string  `json:"nameOrder" dc:"Legacy name sort direction" eg:"asc"`
	PhoneOrder       string  `json:"phoneOrder" dc:"Legacy phone sort direction" eg:"asc"`
	EffectAtOrder    string  `json:"effectAtOrder" dc:"Legacy effectAt sort direction" eg:"desc"`
	ExpireAtOrder    string  `json:"expireAtOrder" dc:"Legacy expireAt sort direction" eg:"desc"`
	GroupIdOrder     string  `json:"groupIdOrder" dc:"Legacy groupId sort direction retained from the old DTO" eg:"asc"`
	PassLevelOrder   string  `json:"passLevelOrder" dc:"Legacy passLevel sort direction" eg:"desc"`
	ContainerIdOrder string  `json:"containerIdOrder" dc:"Legacy containerId sort direction" eg:"asc"`
	StatusOrder      string  `json:"statusOrder" dc:"Legacy status sort direction" eg:"asc"`
	CreatedAtOrder   string  `json:"createdAtOrder" dc:"Legacy createdAt sort direction" eg:"desc"`
	UpdatedAtOrder   string  `json:"updatedAtOrder" dc:"Legacy updatedAt sort direction" eg:"desc"`
	DeletedAtOrder   string  `json:"deletedAtOrder" dc:"Legacy deletedAt sort direction" eg:"desc"`
	CreateByOrder    string  `json:"createByOrder" dc:"Legacy createBy sort direction" eg:"asc"`
	UpdateByOrder    string  `json:"updateByOrder" dc:"Legacy updateBy sort direction" eg:"asc"`
}
