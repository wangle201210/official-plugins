// This file declares CAS ticket validation endpoint DTOs.

package v1

import "github.com/gogf/gf/v2/frame/g"

// CasLoginReq defines a CAS ticket validation request.
type CasLoginReq struct {
	g.Meta `path:"/uidentity/cas/login" method:"post" tags:"UIdentity CAS" summary:"Validate CAS ticket" dc:"Validate a CAS ticket through the plugin-configured CAS service validation URL, resolve the returned work code to a plugin account, enforce application access rules and blacklists, and record a CAS login log."`
	Ticket string `json:"ticket" v:"required" dc:"CAS service ticket" eg:"ST-1-abcdef"`
	UserId int64  `json:"userId" dc:"Optional external user ID forwarded to the CAS validation service" eg:"1"`
	AppId  int64  `json:"appId" dc:"Optional application ID for access-rule and blacklist checks" eg:"1"`
}

// CasLoginRes returns the resolved account number after CAS validation.
type CasLoginRes struct {
	Number    string `json:"number" dc:"Resolved account number" eg:"A001"`
	AccountId int64  `json:"accountId" dc:"Resolved account ID" eg:"1"`
	AppId     int64  `json:"appId" dc:"Application ID used for access checks" eg:"1"`
}
