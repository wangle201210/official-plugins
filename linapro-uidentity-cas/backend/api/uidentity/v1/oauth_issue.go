// This file declares OAuth token issue endpoint DTOs.

package v1

import "github.com/gogf/gf/v2/frame/g"

// OAuthIssueReq defines an OAuth token issue request.
type OAuthIssueReq struct {
	g.Meta      `path:"/uidentity/oauth/tokens" method:"post" tags:"UIdentity CAS" summary:"Issue OAuth token" dc:"Issue an OAuth token record for a plugin account and application, persist the token payload, and record an OAuth authorization log." permission:"uidentity:cas:runtime"`
	AccountId   int64  `json:"accountId" v:"required|min:1" dc:"Account ID" eg:"1"`
	AppId       int64  `json:"appId" v:"required|min:1" dc:"Application ID" eg:"1"`
	RedirectUri string `json:"redirectUri" dc:"OAuth redirect URI" eg:"https://example.com/callback"`
	Scope       string `json:"scope" dc:"OAuth scope string" eg:"profile"`
	TtlSeconds  int64  `json:"ttlSeconds" d:"3600" v:"min:60|max:86400" dc:"Token TTL in seconds, between 60 and 86400" eg:"3600"`
}

// OAuthIssueRes returns issued OAuth token values.
type OAuthIssueRes struct {
	Code      string `json:"code" dc:"Authorization code" eg:"code_abcdef"`
	Access    string `json:"access" dc:"Access token" eg:"access_abcdef"`
	Refresh   string `json:"refresh" dc:"Refresh token" eg:"refresh_abcdef"`
	ExpiredAt *int64 `json:"expiredAt" dc:"Expiration time as Unix timestamp in milliseconds" eg:"1776759600000"`
}
