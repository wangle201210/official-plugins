// This file declares the generic resource create endpoint DTO.

package v1

import "github.com/gogf/gf/v2/frame/g"

// ResourceCreateReq defines the request for creating one UIdentity resource record.
type ResourceCreateReq struct {
	g.Meta   `path:"/uidentity/{resource}" method:"post" tags:"UIdentity CAS" summary:"Create UIdentity resource record" dc:"Create one plugin-owned UIdentity resource record. The body is a resource-specific field object using API field names; unknown fields are ignored." permission:"uidentity:cas:write"`
	Resource string         `json:"resource" v:"required" dc:"Resource name" eg:"accounts"`
	Body     map[string]any `json:"body" dc:"Resource-specific field object" eg:"{\"number\":\"A001\",\"name\":\"Alice\"}"`
}
