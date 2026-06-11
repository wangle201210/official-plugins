// This file implements legacy OAuth authorization-code runtime behavior using
// plugin-owned application, account, token, and OAuth log tables.

package uidentity

import (
	"context"
	"encoding/json"
	"net/url"
	"strings"
	"time"

	"github.com/gogf/gf/v2/database/gdb"

	"lina-core/pkg/bizerr"
	"lina-plugin-linapro-uidentity-cas/backend/internal/dao"
	"lina-plugin-linapro-uidentity-cas/backend/internal/model/do"
	"lina-plugin-linapro-uidentity-cas/backend/internal/model/entity"
)

const (
	oauthGrantTypeAuthorizationCode = "authorization_code"
	oauthGrantTypeRefreshToken      = "refresh_token"
	oauthTokenTypeBearer            = "Bearer"

	oauthKindAuthorizationCode = "oauth_authorization_code"
	oauthKindAccessToken       = "oauth_access_token"

	oauthCodePrefix    = "oauth_code:"
	oauthAccessPrefix  = "oauth_access:"
	oauthRefreshPrefix = "oauth_refresh:"

	oauthCodeTTL   = 5 * time.Minute
	oauthAccessTTL = 2 * time.Hour
	// oauthRefreshTTL mirrors the old go-oauth2 DefaultAuthorizeCodeTokenCfg
	// refresh-token lifetime so refresh grants stay usable after the access
	// token expires.
	oauthRefreshTTL = 72 * time.Hour
)

type oauthRuntimePayload struct {
	Kind        string `json:"kind"`
	Code        string `json:"code,omitempty"`
	AccessToken string `json:"accessToken,omitempty"`
	Refresh     string `json:"refresh,omitempty"`
	AccountID   int64  `json:"accountId"`
	Number      string `json:"number"`
	AppID       int64  `json:"appId"`
	ClientID    string `json:"clientId"`
	RedirectURI string `json:"redirectUri"`
	Scope       string `json:"scope"`
	State       string `json:"state,omitempty"`
	// AccessExpiredAt bounds the access token separately from the stored row,
	// whose expired_at carries the longer refresh-token lifetime.
	AccessExpiredAt int64 `json:"accessExpiredAt,omitempty"`
}

// IssueOAuthAuthorizationCode validates credentials and creates one OAuth code.
func (s *serviceImpl) IssueOAuthAuthorizationCode(ctx context.Context, in OAuthAuthorizationCodeInput) (*OAuthAuthorizationCodeOutput, error) {
	app, err := s.runtimeApplicationByClientID(ctx, in.ClientID)
	if err != nil {
		return nil, err
	}
	redirectURI, err := oauthResolveRedirectURI(app.CallbackUrl, in.RedirectURI)
	if err != nil {
		return nil, err
	}
	account, err := s.getAccountByNumber(ctx, in.Number)
	if err != nil {
		return nil, err
	}
	if err := s.verifyAccountPassword(ctx, account, in.Password); err != nil {
		return nil, err
	}
	if err := s.ensureRuntimeAccess(ctx, account, app); err != nil {
		return nil, err
	}
	code, err := randomToken("OC")
	if err != nil {
		return nil, err
	}
	codeTTL := oauthTTL(in.TtlSeconds, oauthCodeTTL)
	expiredAt := time.Now().Add(codeTTL)
	payload := oauthRuntimePayload{
		Kind:        oauthKindAuthorizationCode,
		Code:        code,
		AccountID:   account.Id,
		Number:      account.Number,
		AppID:       app.Id,
		ClientID:    app.ClientId,
		RedirectURI: redirectURI,
		Scope:       strings.TrimSpace(in.Scope),
		State:       strings.TrimSpace(in.State),
	}
	if err := s.createRuntimeToken(ctx, do.Oauth2Token{
		Code:      oauthCodePrefix + code,
		ExpiredAt: expiredAt.UnixMilli(),
	}, payload); err != nil {
		return nil, err
	}
	redirectURL := callbackWithQuery(redirectURI, "code", code)
	if payload.State != "" {
		redirectURL = callbackWithQuery(redirectURL, "state", payload.State)
	}
	millis := expiredAt.UnixMilli()
	return &OAuthAuthorizationCodeOutput{
		Code:        code,
		RedirectURL: redirectURL,
		ExpiredAt:   &millis,
		State:       payload.State,
	}, nil
}

// ExchangeOAuthAuthorizationCode consumes a code and issues OAuth tokens, or
// rotates an access/refresh pair when the refresh_token grant is requested.
func (s *serviceImpl) ExchangeOAuthAuthorizationCode(ctx context.Context, in OAuthTokenExchangeInput) (*OAuthTokenExchangeOutput, error) {
	if !oauthGrantTypeSupported(in.GrantType) {
		return nil, bizerr.NewCode(CodeOAuthGrantInvalid)
	}
	if strings.TrimSpace(in.GrantType) == oauthGrantTypeRefreshToken {
		return s.refreshOAuthAccessToken(ctx, in)
	}
	app, err := s.runtimeApplicationByClientID(ctx, in.ClientID)
	if err != nil {
		return nil, err
	}
	if !oauthClientSecretMatches(app.SecretKey, in.ClientSecret) {
		return nil, bizerr.NewCode(CodeApplicationSecretInvalid)
	}
	if strings.TrimSpace(in.Code) == "" {
		return nil, bizerr.NewCode(CodeOAuthGrantInvalid)
	}
	token, payload, err := s.oauthAuthorizationCode(ctx, in.Code)
	if err != nil {
		return nil, err
	}
	if payload.ClientID != app.ClientId || payload.AppID != app.Id {
		return nil, bizerr.NewCode(CodeOAuthGrantInvalid)
	}
	if !oauthRedirectExchangeMatches(payload.RedirectURI, in.RedirectURI) {
		return nil, bizerr.NewCode(CodeOAuthRedirectInvalid)
	}
	account, err := s.getAccountByID(ctx, payload.AccountID)
	if err != nil {
		return nil, err
	}
	if err := s.ensureRuntimeAccess(ctx, account, app); err != nil {
		return nil, err
	}
	return s.issueOAuthAccessPair(ctx, token.Id, in.TtlSeconds, oauthRuntimePayload{
		AccountID:   account.Id,
		Number:      account.Number,
		AppID:       app.Id,
		ClientID:    app.ClientId,
		RedirectURI: payload.RedirectURI,
		Scope:       payload.Scope,
		State:       payload.State,
	})
}

// refreshOAuthAccessToken rotates one access/refresh pair like the old
// go-oauth2 default refresh config: the consumed access and refresh tokens are
// removed and a new pair is generated.
func (s *serviceImpl) refreshOAuthAccessToken(ctx context.Context, in OAuthTokenExchangeInput) (*OAuthTokenExchangeOutput, error) {
	app, err := s.runtimeApplicationByClientID(ctx, in.ClientID)
	if err != nil {
		return nil, err
	}
	if !oauthClientSecretMatches(app.SecretKey, in.ClientSecret) {
		return nil, bizerr.NewCode(CodeApplicationSecretInvalid)
	}
	if strings.TrimSpace(in.RefreshToken) == "" {
		return nil, bizerr.NewCode(CodeOAuthGrantInvalid)
	}
	token, payload, err := s.oauthTokenByRefresh(ctx, in.RefreshToken)
	if err != nil {
		return nil, err
	}
	if payload.ClientID != app.ClientId || payload.AppID != app.Id {
		return nil, bizerr.NewCode(CodeOAuthGrantInvalid)
	}
	account, err := s.getAccountByID(ctx, payload.AccountID)
	if err != nil {
		return nil, err
	}
	if err := s.ensureRuntimeAccess(ctx, account, app); err != nil {
		return nil, err
	}
	return s.issueOAuthAccessPair(ctx, token.Id, in.TtlSeconds, oauthRuntimePayload{
		AccountID:   payload.AccountID,
		Number:      payload.Number,
		AppID:       payload.AppID,
		ClientID:    payload.ClientID,
		RedirectURI: payload.RedirectURI,
		Scope:       payload.Scope,
		State:       payload.State,
	})
}

// issueOAuthAccessPair consumes the granting row and stores one access/refresh
// pair whose row outlives the access token by the refresh lifetime.
func (s *serviceImpl) issueOAuthAccessPair(ctx context.Context, consumedTokenID int64, ttlSeconds int64, payload oauthRuntimePayload) (*OAuthTokenExchangeOutput, error) {
	ttl := oauthTTL(ttlSeconds, oauthAccessTTL)
	access, err := randomToken("OA")
	if err != nil {
		return nil, err
	}
	refresh, err := randomToken("OR")
	if err != nil {
		return nil, err
	}
	now := time.Now()
	accessExpiredAt := now.Add(ttl)
	rowExpiredAt := now.Add(oauthRefreshTTL)
	if accessExpiredAt.After(rowExpiredAt) {
		rowExpiredAt = accessExpiredAt
	}
	payload.Kind = oauthKindAccessToken
	payload.AccessToken = access
	payload.Refresh = refresh
	payload.AccessExpiredAt = accessExpiredAt.UnixMilli()
	if err := s.consumeOAuthGrantAndCreateAccess(ctx, consumedTokenID, rowExpiredAt, access, refresh, payload); err != nil {
		return nil, err
	}
	millis := accessExpiredAt.UnixMilli()
	return &OAuthTokenExchangeOutput{
		AccessToken:  access,
		RefreshToken: refresh,
		TokenType:    oauthTokenTypeBearer,
		ExpiresIn:    int64(ttl.Seconds()),
		ExpiredAt:    &millis,
		Scope:        payload.Scope,
	}, nil
}

// GetOAuthAccessTokenInfo returns OAuth token-bound user information.
func (s *serviceImpl) GetOAuthAccessTokenInfo(ctx context.Context, accessToken string) (*OAuthAccessTokenInfoOutput, error) {
	payload, err := s.oauthAccessToken(ctx, accessToken)
	if err != nil {
		return nil, err
	}
	account, err := s.getAccountByID(ctx, payload.AccountID)
	if err != nil {
		return nil, err
	}
	app, err := s.runtimeApplication(ctx, payload.AppID)
	if err != nil {
		return nil, err
	}
	if err := s.ensureRuntimeAccess(ctx, account, app); err != nil {
		return nil, err
	}
	user, err := s.runtimeAccountProjection(ctx, account)
	if err != nil {
		return nil, err
	}
	return &OAuthAccessTokenInfoOutput{
		User:  user,
		App:   runtimeApplicationProjection(app),
		Scope: payload.Scope,
	}, nil
}

func (s *serviceImpl) oauthAuthorizationCode(ctx context.Context, code string) (*entity.Oauth2Token, *oauthRuntimePayload, error) {
	return s.oauthTokenByCode(ctx, oauthCodePrefix+strings.TrimSpace(code), oauthKindAuthorizationCode)
}

func (s *serviceImpl) oauthAccessToken(ctx context.Context, accessToken string) (*oauthRuntimePayload, error) {
	var token *entity.Oauth2Token
	err := dao.Oauth2Token.Ctx(ctx).
		Where(dao.Oauth2Token.Columns().Access, oauthAccessPrefix+strings.TrimSpace(accessToken)).
		Scan(&token)
	if err != nil {
		return nil, err
	}
	_, payload, err := parseOAuthRuntimeToken(token, oauthKindAccessToken)
	if err != nil {
		return nil, err
	}
	if oauthAccessPayloadExpired(payload, time.Now()) {
		return nil, bizerr.NewCode(CodeTicketInvalid)
	}
	return payload, nil
}

func (s *serviceImpl) oauthTokenByRefresh(ctx context.Context, refreshToken string) (*entity.Oauth2Token, *oauthRuntimePayload, error) {
	var token *entity.Oauth2Token
	err := dao.Oauth2Token.Ctx(ctx).
		Where(dao.Oauth2Token.Columns().Refresh, oauthRefreshPrefix+strings.TrimSpace(refreshToken)).
		Scan(&token)
	if err != nil {
		return nil, nil, err
	}
	return parseOAuthRuntimeToken(token, oauthKindAccessToken)
}

func (s *serviceImpl) oauthTokenByCode(ctx context.Context, code string, kind string) (*entity.Oauth2Token, *oauthRuntimePayload, error) {
	var token *entity.Oauth2Token
	err := dao.Oauth2Token.Ctx(ctx).
		Where(dao.Oauth2Token.Columns().Code, code).
		Scan(&token)
	if err != nil {
		return nil, nil, err
	}
	return parseOAuthRuntimeToken(token, kind)
}

func (s *serviceImpl) consumeOAuthGrantAndCreateAccess(
	ctx context.Context,
	consumedTokenID int64,
	expiredAt time.Time,
	access string,
	refresh string,
	payload oauthRuntimePayload,
) error {
	content, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	actorID := s.actorID(ctx)
	return dao.Oauth2Token.Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {
		result, err := tx.Model(dao.Oauth2Token.Table()).Safe().Ctx(ctx).
			Where(dao.Oauth2Token.Columns().Id, consumedTokenID).
			Delete()
		if err != nil {
			return err
		}
		affected, err := result.RowsAffected()
		if err != nil {
			return err
		}
		if affected == 0 {
			return bizerr.NewCode(CodeOAuthGrantInvalid)
		}
		if _, err := tx.Model(dao.Oauth2Token.Table()).Safe().Ctx(ctx).
			Data(do.Oauth2Token{
				ExpiredAt: expiredAt.UnixMilli(),
				Code:      oauthAccessPrefix + access,
				Access:    oauthAccessPrefix + access,
				Refresh:   oauthRefreshPrefix + refresh,
				Data:      string(content),
			}).
			Insert(); err != nil {
			return err
		}
		_, err = tx.Model(dao.OauthLog.Table()).Safe().Ctx(ctx).
			Data(do.OauthLog{
				UserId:      payload.AccountID,
				AppId:       payload.AppID,
				RedirectUri: payload.RedirectURI,
				Scope:       payload.Scope,
				CreateBy:    actorID,
				UpdateBy:    actorID,
			}).
			Insert()
		return err
	})
}

func parseOAuthRuntimeToken(token *entity.Oauth2Token, kind string) (*entity.Oauth2Token, *oauthRuntimePayload, error) {
	// gf Scan leaves the pointer nil when no row matches.
	if token == nil || runtimeTokenExpired(token.ExpiredAt, time.Now()) {
		return nil, nil, bizerr.NewCode(CodeTicketInvalid)
	}
	payload := &oauthRuntimePayload{}
	if err := json.Unmarshal([]byte(token.Data), payload); err != nil {
		return nil, nil, bizerr.NewCode(CodeTicketInvalid)
	}
	if payload.Kind != kind {
		return nil, nil, bizerr.NewCode(CodeTicketInvalid)
	}
	return token, payload, nil
}

func oauthTTL(ttlSeconds int64, fallback time.Duration) time.Duration {
	if ttlSeconds <= 0 {
		return fallback
	}
	return time.Duration(ttlSeconds) * time.Second
}

func oauthGrantTypeSupported(grantType string) bool {
	trimmed := strings.TrimSpace(grantType)
	return trimmed == "" || trimmed == oauthGrantTypeAuthorizationCode || trimmed == oauthGrantTypeRefreshToken
}

// oauthAccessPayloadExpired reports whether the payload-level access expiry
// passed; rows created before AccessExpiredAt existed rely on row expiry only.
func oauthAccessPayloadExpired(payload *oauthRuntimePayload, now time.Time) bool {
	return payload != nil && payload.AccessExpiredAt > 0 && payload.AccessExpiredAt <= now.UnixMilli()
}

func oauthClientSecretMatches(expected string, actual string) bool {
	trimmedExpected := strings.TrimSpace(expected)
	trimmedActual := strings.TrimSpace(actual)
	if trimmedExpected == "" || trimmedActual == "" {
		return false
	}
	return trimmedActual == trimmedExpected || trimmedActual == url.QueryEscape(trimmedExpected)
}

func oauthResolveRedirectURI(callbackURL string, requestedURI string) (string, error) {
	callback := strings.TrimSpace(callbackURL)
	requested := strings.TrimSpace(requestedURI)
	if requested == "" {
		requested = callback
	}
	if requested == "" {
		return "", nil
	}
	if !oauthURLValid(requested) {
		return "", bizerr.NewCode(CodeOAuthRedirectInvalid)
	}
	if callback == "" {
		return requested, nil
	}
	if !oauthRedirectMatches(callback, requested) {
		return "", bizerr.NewCode(CodeOAuthRedirectInvalid)
	}
	return requested, nil
}

func oauthRedirectExchangeMatches(storedURI string, requestedURI string) bool {
	stored := strings.TrimSpace(storedURI)
	requested := strings.TrimSpace(requestedURI)
	return requested == "" || stored == "" || requested == stored
}

func oauthRedirectMatches(callbackURL string, requestedURI string) bool {
	callback, callbackErr := url.Parse(strings.TrimSpace(callbackURL))
	requested, requestedErr := url.Parse(strings.TrimSpace(requestedURI))
	if callbackErr != nil || requestedErr != nil || callback.Scheme == "" || callback.Host == "" || requested.Scheme == "" || requested.Host == "" {
		return strings.TrimSpace(callbackURL) == strings.TrimSpace(requestedURI)
	}
	return strings.EqualFold(callback.Scheme, requested.Scheme) &&
		strings.EqualFold(callback.Host, requested.Host) &&
		callback.Path == requested.Path
}

func oauthURLValid(rawURL string) bool {
	parsed, err := url.Parse(strings.TrimSpace(rawURL))
	if err != nil {
		return false
	}
	return parsed.Scheme != "" && parsed.Host != ""
}
