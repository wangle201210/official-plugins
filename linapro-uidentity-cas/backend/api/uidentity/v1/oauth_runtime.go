// This file declares legacy-compatible OAuth authorization-code runtime DTOs.

package v1

import "github.com/gogf/gf/v2/frame/g"

// OAuthAuthorizationCodeReq defines password-backed authorization-code issue.
type OAuthAuthorizationCodeReq struct {
	g.Meta      `path:"/uidentity/oauth/authorization-codes" method:"post" tags:"UIdentity OAuth Runtime" summary:"Issue OAuth authorization code" dc:"Validate application client ID, redirect URI, account password and access rules, then issue a one-time OAuth authorization code without rendering legacy login pages."`
	ClientId    string `json:"clientId" v:"required" dc:"Application client ID" eg:"portal"`
	RedirectUri string `json:"redirectUri" dc:"Requested OAuth redirect URI; defaults to the application callback URL when empty" eg:"https://example.com/oauth/callback"`
	Scope       string `json:"scope" dc:"OAuth scope granted to the authorization code" eg:"read_user_info"`
	State       string `json:"state" dc:"Opaque client state returned in the redirect URL" eg:"csrf-state"`
	Number      string `json:"number" v:"required" dc:"Account number" eg:"A001"`
	Password    string `json:"password" v:"required" dc:"Plaintext password submitted by the runtime client" eg:"S3cure@2026"`
	TtlSeconds  int64  `json:"ttlSeconds" d:"300" v:"min:60|max:1800" dc:"Authorization code TTL in seconds, between 60 and 1800" eg:"300"`
}

// OAuthAccessTokenReq defines authorization-code token exchange.
type OAuthAccessTokenReq struct {
	g.Meta       `path:"/uidentity/oauth/access-tokens" method:"post" tags:"UIdentity OAuth Runtime" summary:"Exchange OAuth authorization code" dc:"Validate client secret and a one-time authorization code, consume the code, issue access and refresh tokens, and record an OAuth authorization log."`
	GrantType    string `json:"grantType" d:"authorization_code" dc:"OAuth grant type; only authorization_code is supported" eg:"authorization_code"`
	ClientId     string `json:"clientId" v:"required" dc:"Application client ID" eg:"portal"`
	ClientSecret string `json:"clientSecret" v:"required" dc:"Application client secret, raw or URL-escaped for legacy clients" eg:"secret"`
	Code         string `json:"code" v:"required" dc:"One-time authorization code" eg:"OC_abcdef"`
	RedirectUri  string `json:"redirectUri" dc:"Redirect URI used by the authorization request; required to match when the code stored one" eg:"https://example.com/oauth/callback"`
	TtlSeconds   int64  `json:"ttlSeconds" d:"7200" v:"min:60|max:86400" dc:"Access token TTL in seconds, between 60 and 86400" eg:"7200"`
}

// OAuthAccessTokenInfoReq defines access-token user-info lookup.
type OAuthAccessTokenInfoReq struct {
	g.Meta      `path:"/uidentity/oauth/access-tokens/{accessToken}/user-info" method:"get" tags:"UIdentity OAuth Runtime" summary:"Get user info by OAuth access token" dc:"Validate an OAuth access token and return the bound account and application projections without consuming the token."`
	AccessToken string `json:"accessToken" v:"required" dc:"OAuth access token" eg:"OA_abcdef"`
}

// OAuthAuthorizationCodeRes returns one-time code data.
type OAuthAuthorizationCodeRes struct {
	Code        string `json:"code" dc:"One-time OAuth authorization code" eg:"OC_abcdef"`
	RedirectUrl string `json:"redirectUrl" dc:"Redirect URL with code and state query parameters when an application callback URL is configured" eg:"https://example.com/oauth/callback?code=OC_abcdef&state=csrf-state"`
	ExpiredAt   *int64 `json:"expiredAt" dc:"Authorization code expiration time as Unix timestamp in milliseconds" eg:"1776759600000"`
	State       string `json:"state" dc:"Opaque client state echoed from the request" eg:"csrf-state"`
}

// OAuthAccessTokenRes returns exchanged token data.
type OAuthAccessTokenRes struct {
	AccessToken  string `json:"accessToken" dc:"OAuth access token" eg:"OA_abcdef"`
	RefreshToken string `json:"refreshToken" dc:"OAuth refresh token" eg:"OR_abcdef"`
	TokenType    string `json:"tokenType" dc:"Token type for Authorization header usage" eg:"Bearer"`
	ExpiresIn    int64  `json:"expiresIn" dc:"Access token lifetime in seconds" eg:"7200"`
	ExpiredAt    *int64 `json:"expiredAt" dc:"Access token expiration time as Unix timestamp in milliseconds" eg:"1776759600000"`
	Scope        string `json:"scope" dc:"Granted OAuth scope" eg:"read_user_info"`
}

// OAuthAccessTokenInfoRes returns token-bound user information.
type OAuthAccessTokenInfoRes struct {
	User  *RuntimeAccount     `json:"user" dc:"Account projection bound to the OAuth access token" eg:"{}"`
	App   *RuntimeApplication `json:"app" dc:"Application projection bound to the OAuth access token" eg:"{}"`
	Scope string              `json:"scope" dc:"Granted OAuth scope" eg:"read_user_info"`
}
