// This file projects plugin-scoped static configuration into the legacy admin
// configuration shapes without reading host-global config or page assets.

package uidentity

import "context"

const (
	configKeyLegacyCASLoginAddr   = "legacy.static.cas.loginAddr"
	configKeyLegacyCASLogoutAddr  = "legacy.static.cas.logoutAddr"
	configKeyLegacyCASRestAddr    = "legacy.static.cas.restAddr"
	configKeyLegacyCASDocs        = "legacy.static.cas.docs"
	configKeyLegacyLDAPVersion    = "legacy.static.ldap.version"
	configKeyLegacyLDAPAddr       = "legacy.static.ldap.addr"
	configKeyLegacyLDAPDocs       = "legacy.static.ldap.docs"
	configKeyLegacyOAuthAuthorize = "legacy.static.oauth.authorization"
	configKeyLegacyOAuthToken     = "legacy.static.oauth.getTokenAddr"
	configKeyLegacyOAuthUserInfo  = "legacy.static.oauth.userInfoAddr"
	configKeyLegacyOAuthLogout    = "legacy.static.oauth.logoutAddr"
	configKeyLegacyOAuthPing      = "legacy.static.oauth.pingAddr"
	configKeyLegacyOAuthDocs      = "legacy.static.oauth.docs"
	configKeyLegacyTokenGet       = "legacy.static.token.getAddr"
	configKeyLegacyTokenCheck     = "legacy.static.token.checkAddr"
	configKeyLegacyTokenDocs      = "legacy.static.token.tokenDocs"
	configKeyLegacyTokenCASDocs   = "legacy.static.token.casDocs"
	configKeyLegacyDefaultAppID   = "legacy.sso.defaultAppid"
	configKeyLegacySSOLogin       = "legacy.sso.loginRedirect"
	configKeyLegacySSOLogout      = "legacy.sso.logoutRedirect"

	defaultLegacyCASLoginAddr   = "/api/v1/uidentity/cas/password-logins"
	defaultLegacyCASLogoutAddr  = "/api/v1/uidentity/cas/tickets/{ticket}"
	defaultLegacyCASRestAddr    = "/api/v1/uidentity/legacy/cas/service-validations.xml"
	defaultLegacyLDAPVersion    = "unsupported"
	defaultLegacyOAuthAuthorize = "/api/v1/uidentity/oauth/authorization-codes"
	defaultLegacyOAuthToken     = "/api/v1/uidentity/oauth/access-tokens"
	defaultLegacyOAuthUserInfo  = "/api/v1/uidentity/oauth/access-tokens/{accessToken}/user-info"
	defaultLegacyOAuthPing      = "/api/v1/uidentity/legacy/health"
	defaultLegacyTokenGet       = "/api/v1/uidentity/runtime-tokens"
	defaultLegacyTokenCheck     = "/api/v1/uidentity/runtime-tokens/{accessToken}/user-info"
)

// LegacyCASConfig returns legacy CAS endpoint metadata from plugin config.
func (s *serviceImpl) LegacyCASConfig(ctx context.Context) (*LegacyCASConfigOutput, error) {
	loginAddr, err := s.configSvc.String(ctx, configKeyLegacyCASLoginAddr, defaultLegacyCASLoginAddr)
	if err != nil {
		return nil, err
	}
	logoutAddr, err := s.configSvc.String(ctx, configKeyLegacyCASLogoutAddr, defaultLegacyCASLogoutAddr)
	if err != nil {
		return nil, err
	}
	restAddr, err := s.configSvc.String(ctx, configKeyLegacyCASRestAddr, defaultLegacyCASRestAddr)
	if err != nil {
		return nil, err
	}
	docs, err := s.configSvc.String(ctx, configKeyLegacyCASDocs, "")
	if err != nil {
		return nil, err
	}
	return &LegacyCASConfigOutput{
		LoginAddr:  loginAddr,
		LogoutAddr: logoutAddr,
		RestAddr:   restAddr,
		Docs:       docs,
	}, nil
}

// LegacyLDAPConfig returns legacy LDAP endpoint metadata from plugin config.
func (s *serviceImpl) LegacyLDAPConfig(ctx context.Context) (*LegacyLDAPConfigOutput, error) {
	version, err := s.configSvc.String(ctx, configKeyLegacyLDAPVersion, defaultLegacyLDAPVersion)
	if err != nil {
		return nil, err
	}
	addr, err := s.configSvc.String(ctx, configKeyLegacyLDAPAddr, "")
	if err != nil {
		return nil, err
	}
	docs, err := s.configSvc.String(ctx, configKeyLegacyLDAPDocs, "")
	if err != nil {
		return nil, err
	}
	return &LegacyLDAPConfigOutput{
		Version: version,
		Addr:    addr,
		Docs:    docs,
	}, nil
}

// LegacyOAuthConfig returns legacy OAuth endpoint metadata from plugin config.
func (s *serviceImpl) LegacyOAuthConfig(ctx context.Context) (*LegacyOAuthConfigOutput, error) {
	authorization, err := s.configSvc.String(ctx, configKeyLegacyOAuthAuthorize, defaultLegacyOAuthAuthorize)
	if err != nil {
		return nil, err
	}
	getTokenAddr, err := s.configSvc.String(ctx, configKeyLegacyOAuthToken, defaultLegacyOAuthToken)
	if err != nil {
		return nil, err
	}
	userInfoAddr, err := s.configSvc.String(ctx, configKeyLegacyOAuthUserInfo, defaultLegacyOAuthUserInfo)
	if err != nil {
		return nil, err
	}
	logoutAddr, err := s.configSvc.String(ctx, configKeyLegacyOAuthLogout, "")
	if err != nil {
		return nil, err
	}
	pingAddr, err := s.configSvc.String(ctx, configKeyLegacyOAuthPing, defaultLegacyOAuthPing)
	if err != nil {
		return nil, err
	}
	docs, err := s.configSvc.String(ctx, configKeyLegacyOAuthDocs, "")
	if err != nil {
		return nil, err
	}
	return &LegacyOAuthConfigOutput{
		Authorization: authorization,
		GetTokenAddr:  getTokenAddr,
		UserInfoAddr:  userInfoAddr,
		LogoutAddr:    logoutAddr,
		PingAddr:      pingAddr,
		Docs:          docs,
	}, nil
}

// LegacyTokenConfig returns legacy runtime token endpoint metadata from config.
func (s *serviceImpl) LegacyTokenConfig(ctx context.Context) (*LegacyTokenConfigOutput, error) {
	getAddr, err := s.configSvc.String(ctx, configKeyLegacyTokenGet, defaultLegacyTokenGet)
	if err != nil {
		return nil, err
	}
	checkAddr, err := s.configSvc.String(ctx, configKeyLegacyTokenCheck, defaultLegacyTokenCheck)
	if err != nil {
		return nil, err
	}
	tokenDocs, err := s.configSvc.String(ctx, configKeyLegacyTokenDocs, "")
	if err != nil {
		return nil, err
	}
	casDocs, err := s.configSvc.String(ctx, configKeyLegacyTokenCASDocs, "")
	if err != nil {
		return nil, err
	}
	return &LegacyTokenConfigOutput{
		GetAddr:   getAddr,
		CheckAddr: checkAddr,
		TokenDocs: tokenDocs,
		CasDocs:   casDocs,
	}, nil
}

// LegacyRedirectConfig returns old redirect shell configuration from plugin config.
func (s *serviceImpl) LegacyRedirectConfig(ctx context.Context) (*LegacyRedirectConfigOutput, error) {
	defaultAppID, err := s.configSvc.String(ctx, configKeyLegacyDefaultAppID, "")
	if err != nil {
		return nil, err
	}
	ssoLogin, err := s.configSvc.String(ctx, configKeyLegacySSOLogin, "")
	if err != nil {
		return nil, err
	}
	ssoLogout, err := s.configSvc.String(ctx, configKeyLegacySSOLogout, "")
	if err != nil {
		return nil, err
	}
	unionIDBind, err := s.configSvc.String(ctx, configKeyUnionIDBindCallbackURL, "")
	if err != nil {
		return nil, err
	}
	wechatLogin, err := s.configSvc.String(ctx, configKeyWechatLoginRedirectURL, "")
	if err != nil {
		return nil, err
	}
	activationRedirect, err := s.configSvc.String(ctx, configKeyActivationWechatRedirectURL, "")
	if err != nil {
		return nil, err
	}
	return &LegacyRedirectConfigOutput{
		DefaultAppID:          defaultAppID,
		SSOLoginRedirect:      ssoLogin,
		SSOLogoutRedirect:     ssoLogout,
		UnionIDBindRedirect:   unionIDBind,
		WechatLoginRedirect:   wechatLogin,
		ActivationRedirectURL: activationRedirect,
	}, nil
}
