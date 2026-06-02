// This file declares the old account detail endpoint DTO.

package v1

import "github.com/gogf/gf/v2/frame/g"

// ResourceGetReq defines the old GET /api/v1/account/{id} request.
type ResourceGetReq struct {
	g.Meta `path:"/api/v1/account/{id}" method:"get" tags:"UIdentity CAS" summary:"Get account" dc:"Match the old uidentity/admin account detail contract." permission:"uidentity:cas:read"`
	Id     int64 `json:"id" v:"required|min:1" dc:"Account ID from the legacy path parameter" eg:"1"`
}
