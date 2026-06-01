// This file declares the legacy CAS XML validation endpoint DTO used by old
// CAS clients that expect the classic serviceResponse document.

package v1

import "github.com/gogf/gf/v2/frame/g"

// LegacyCASServiceValidateXMLReq defines XML service-ticket validation.
type LegacyCASServiceValidateXMLReq struct {
	g.Meta `path:"/uidentity/legacy/cas/service-validations.xml" method:"post" tags:"UIdentity Legacy CAS Runtime" summary:"Validate CAS service ticket as XML" dc:"Consume and validate a CAS service ticket, then write a CAS serviceResponse XML document compatible with the old admin serviceValidate endpoint."`
	Ticket string `json:"ticket" v:"required" dc:"CAS service ticket" eg:"ST_abcdef"`
	UserId int64  `json:"userId" dc:"Optional selected account ID to validate delegated access" eg:"1"`
}

// LegacyCASServiceValidateXMLRes is empty because the controller writes XML.
type LegacyCASServiceValidateXMLRes struct{}
