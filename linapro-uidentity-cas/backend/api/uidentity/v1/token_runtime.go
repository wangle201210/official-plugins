// This file declares legacy-compatible runtime token endpoint DTOs.

package v1

import "github.com/gogf/gf/v2/frame/g"

// RuntimeTokenIssueReq defines account-password runtime token issue.
type RuntimeTokenIssueReq struct {
	g.Meta   `path:"/uidentity/runtime-tokens" method:"post" tags:"UIdentity Token Runtime" summary:"Issue runtime access token" dc:"Validate application secret, account password and access rules, then issue a short-lived runtime access token for non-page clients."`
	ClientId string `json:"clientId" v:"required" dc:"Application client ID" eg:"portal"`
	Secret   string `json:"secret" v:"required" dc:"Application secret key" eg:"secret"`
	Number   string `json:"number" v:"required" dc:"Account number" eg:"A001"`
	Password string `json:"password" v:"required" dc:"Plaintext password submitted by the runtime client" eg:"S3cure@2026"`
}

// RuntimeTokenInfoReq defines token user-info lookup.
type RuntimeTokenInfoReq struct {
	g.Meta      `path:"/uidentity/runtime-tokens/{accessToken}/user-info" method:"get" tags:"UIdentity Token Runtime" summary:"Get user info by runtime token" dc:"Validate a runtime access token and return the primary and delegated account projections available for the token application."`
	AccessToken string `json:"accessToken" v:"required" dc:"Runtime access token" eg:"AT_abcdef"`
}

// RuntimeTokenIssueRes returns issued runtime access token.
type RuntimeTokenIssueRes struct {
	AccessToken string `json:"accessToken" dc:"Runtime access token" eg:"AT_abcdef"`
	ExpiredAt   *int64 `json:"expiredAt" dc:"Expiration time as Unix timestamp in milliseconds" eg:"1776759600000"`
}

// RuntimeTokenInfoRes returns token-bound user information.
type RuntimeTokenInfoRes struct {
	User  *RuntimeAccount   `json:"user" dc:"Primary account projection" eg:"{}"`
	Users []*RuntimeAccount `json:"users" dc:"Accounts accessible for the token application" eg:"[]"`
	App   *RuntimeApplication `json:"app" dc:"Application projection" eg:"{}"`
}
