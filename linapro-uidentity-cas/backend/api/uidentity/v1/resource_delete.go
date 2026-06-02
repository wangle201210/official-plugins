// This file declares the old account delete endpoint DTO.

package v1

import "github.com/gogf/gf/v2/frame/g"

// ResourceDeleteReq defines the old DELETE /api/v1/account request.
type ResourceDeleteReq struct {
	g.Meta `path:"/api/v1/account" method:"delete" tags:"UIdentity CAS" summary:"Delete accounts" dc:"Match the old uidentity/admin account delete contract. The request body carries ids as an array." permission:"uidentity:cas:delete"`
	Ids    []int64 `json:"ids" v:"required" dc:"Account IDs" eg:"[1,2,3]"`
}
