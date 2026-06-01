// This file declares the generic resource update endpoint DTO.

package v1

import "github.com/gogf/gf/v2/frame/g"

// ResourceUpdateReq defines the request for updating one UIdentity resource record.
type ResourceUpdateReq struct {
	g.Meta   `path:"/uidentity/{resource}/{id}" method:"put" tags:"UIdentity CAS" summary:"Update UIdentity resource record" dc:"Update one plugin-owned UIdentity resource record by ID. The body is a partial resource-specific field object using API field names; omitted fields remain unchanged." permission:"uidentity:cas:write"`
	Resource string         `json:"resource" v:"required" dc:"Resource name" eg:"accounts"`
	Id       int64          `json:"id" v:"required|min:1" dc:"Record ID" eg:"1"`
	Body     map[string]any `json:"body" dc:"Partial resource-specific field object" eg:"{\"name\":\"Alice Zhang\"}"`
}
