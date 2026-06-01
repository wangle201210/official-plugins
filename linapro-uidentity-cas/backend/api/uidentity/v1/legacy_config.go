// This file declares legacy static configuration discovery DTOs for old admin
// clients that read CAS, LDAP, OAuth, and token endpoint metadata.

package v1

import "github.com/gogf/gf/v2/frame/g"

// LegacyCASConfigReq defines the legacy CAS static config lookup.
type LegacyCASConfigReq struct {
	g.Meta `path:"/uidentity/legacy/config/cas" method:"get" tags:"UIdentity Legacy Configuration" summary:"Get legacy CAS configuration" dc:"Return plugin-scoped CAS endpoint metadata compatible with the old admin static configuration response." permission:"uidentity:cas:read"`
}

// LegacyLDAPConfigReq defines the legacy LDAP static config lookup.
type LegacyLDAPConfigReq struct {
	g.Meta `path:"/uidentity/legacy/config/ldap" method:"get" tags:"UIdentity Legacy Configuration" summary:"Get legacy LDAP configuration" dc:"Return plugin-scoped LDAP endpoint metadata compatible with the old admin static configuration response. The plugin does not start an LDAP executor by default." permission:"uidentity:cas:read"`
}

// LegacyOAuthConfigReq defines the legacy OAuth static config lookup.
type LegacyOAuthConfigReq struct {
	g.Meta `path:"/uidentity/legacy/config/oauth" method:"get" tags:"UIdentity Legacy Configuration" summary:"Get legacy OAuth configuration" dc:"Return plugin-scoped OAuth endpoint metadata compatible with the old admin static configuration response." permission:"uidentity:cas:read"`
}

// LegacyTokenConfigReq defines the legacy runtime token static config lookup.
type LegacyTokenConfigReq struct {
	g.Meta `path:"/uidentity/legacy/config/token" method:"get" tags:"UIdentity Legacy Configuration" summary:"Get legacy token configuration" dc:"Return plugin-scoped runtime token endpoint metadata compatible with the old admin static configuration response." permission:"uidentity:cas:read"`
}

// LegacyCASConfigRes returns legacy CAS endpoint metadata.
type LegacyCASConfigRes struct {
	LoginAddr  string `json:"LoginAddr" dc:"CAS login endpoint address shown to old admin clients" eg:"/api/v1/uidentity/cas/password-logins"`
	LogoutAddr string `json:"LogoutAddr" dc:"CAS logout endpoint address shown to old admin clients" eg:"/api/v1/uidentity/cas/tickets/{ticket}"`
	RestAddr   string `json:"RestAddr" dc:"CAS service-validation endpoint address shown to old admin clients" eg:"/api/v1/uidentity/legacy/cas/service-validations.xml"`
	Docs       string `json:"Docs" dc:"CAS integration documentation URL or text pointer" eg:"https://example.com/docs/cas"`
}

// LegacyLDAPConfigRes returns legacy LDAP endpoint metadata.
type LegacyLDAPConfigRes struct {
	Version string `json:"Version" dc:"LDAP compatibility version label shown to old admin clients" eg:"unsupported"`
	Addr    string `json:"Addr" dc:"LDAP service endpoint address shown to old admin clients" eg:""`
	Docs    string `json:"Docs" dc:"LDAP integration documentation URL or text pointer" eg:"https://example.com/docs/ldap"`
}

// LegacyOAuthConfigRes returns legacy OAuth endpoint metadata.
type LegacyOAuthConfigRes struct {
	Authorization string `json:"Authorization" dc:"OAuth authorization-code issue endpoint address shown to old admin clients" eg:"/api/v1/uidentity/oauth/authorization-codes"`
	GetTokenAddr  string `json:"GetTokenAddr" dc:"OAuth access-token exchange endpoint address shown to old admin clients" eg:"/api/v1/uidentity/oauth/access-tokens"`
	UserInfoAddr  string `json:"UserInfoAddr" dc:"OAuth user-info endpoint address shown to old admin clients" eg:"/api/v1/uidentity/oauth/access-tokens/{accessToken}/user-info"`
	LogoutAddr    string `json:"LogoutAddr" dc:"OAuth logout endpoint address shown to old admin clients" eg:""`
	PingAddr      string `json:"PingAddr" dc:"OAuth ping endpoint address shown to old admin clients" eg:"/api/v1/uidentity/legacy/health"`
	Docs          string `json:"Docs" dc:"OAuth integration documentation URL or text pointer" eg:"https://example.com/docs/oauth"`
}

// LegacyTokenConfigRes returns legacy runtime token endpoint metadata.
type LegacyTokenConfigRes struct {
	GetAddr   string `json:"GetAddr" dc:"Runtime token issue endpoint address shown to old admin clients" eg:"/api/v1/uidentity/runtime-tokens"`
	CheckAddr string `json:"CheckAddr" dc:"Runtime token user-info endpoint address shown to old admin clients" eg:"/api/v1/uidentity/runtime-tokens/{accessToken}/user-info"`
	TokenDocs string `json:"TokenDocs" dc:"Runtime token integration documentation URL or text pointer" eg:"https://example.com/docs/token"`
	CasDocs   string `json:"CasDocs" dc:"CAS integration documentation URL or text pointer" eg:"https://example.com/docs/cas"`
}
