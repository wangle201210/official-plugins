// This file declares the generic resource detail endpoint DTO.

package v1

import "github.com/gogf/gf/v2/frame/g"

// ResourceGetReq defines the request for reading one UIdentity resource record.
type ResourceGetReq struct {
	g.Meta   `path:"/uidentity/{resource}/{id}" method:"get" tags:"UIdentity CAS" summary:"Get UIdentity resource record" dc:"Read one plugin-owned UIdentity resource record by ID, including stable related-name projections where available." permission:"uidentity:cas:read"`
	Resource string `json:"resource" v:"required" dc:"Resource name" eg:"accounts"`
	Id       int64  `json:"id" v:"required|min:1" dc:"Record ID" eg:"1"`
}
