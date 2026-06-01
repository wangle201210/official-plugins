// This file declares the generic resource delete endpoint DTO.

package v1

import "github.com/gogf/gf/v2/frame/g"

// ResourceDeleteReq defines the request for deleting one or more UIdentity resource records.
type ResourceDeleteReq struct {
	g.Meta   `path:"/uidentity/{resource}/{ids}" method:"delete" tags:"UIdentity CAS" summary:"Delete UIdentity resource records" dc:"Delete one or more plugin-owned UIdentity resource records. IDs are comma-separated and capped at 100 per request." permission:"uidentity:cas:delete"`
	Resource string `json:"resource" v:"required" dc:"Resource name" eg:"accounts"`
	Ids      string `json:"ids" v:"required" dc:"Comma-separated record IDs, max 100" eg:"1,2,3"`
}
