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
	plugincontract "lina-core/pkg/plugin/capability/contract"
	"lina-plugin-linapro-uidentity-cas/backend/internal/dao"
	"lina-plugin-linapro-uidentity-cas/backend/internal/model/do"
	"lina-plugin-linapro-uidentity-cas/backend/internal/model/entity"
)

const (
	oauthGrantTypeAuthorizationCode = "authorization_code"
	oauthTokenTypeBearer            = "Bearer"

	oauthKindAuthorizationCode = "oauth_authorization_code"
	oauthKindAccessToken       = "oauth_access_token"

	oauthCodePrefix    = "oauth_code:"
	oauthAccessPrefix  = "oauth_access:"
	oauthRefreshPrefix = "oauth_refresh:"

	oauthCodeTTL   = 5 * time.Minute
	oauthAccessTTL = 2 * time.Hour
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
	if !passwordMatches(account, in.Password) {
		return nil, bizerr.NewCode(CodeInvalidCredentials)
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
	if err := s.createRuntimeToken(ctx, do.OauthToken{
		Code:      oauthCodePrefix + code,
		ExpiredAt: &expiredAt,
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

// ExchangeOAuthAuthorizationCode consumes a code and issues OAuth tokens.
func (s *serviceImpl) ExchangeOAuthAuthorizationCode(ctx context.Context, in OAuthTokenExchangeInput) (*OAuthTokenExchangeOutput, error) {
	if !oauthGrantTypeSupported(in.GrantType) {
		return nil, bizerr.NewCode(CodeOAuthGrantInvalid)
	}
	app, err := s.runtimeApplicationByClientID(ctx, in.ClientID)
	if err != nil {
		return nil, err
	}
	if !oauthClientSecretMatches(app.SecretKey, in.ClientSecret) {
		return nil, bizerr.NewCode(CodeApplicationSecretInvalid)
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
	ttl := oauthTTL(in.TtlSeconds, oauthAccessTTL)
	access, err := randomToken("OA")
	if err != nil {
		return nil, err
	}
	refresh, err := randomToken("OR")
	if err != nil {
		return nil, err
	}
	expiredAt := time.Now().Add(ttl)
	accessPayload := oauthRuntimePayload{
		Kind:        oauthKindAccessToken,
		AccessToken: access,
		Refresh:     refresh,
		AccountID:   account.Id,
		Number:      account.Number,
		AppID:       app.Id,
		ClientID:    app.ClientId,
		RedirectURI: payload.RedirectURI,
		Scope:       payload.Scope,
		State:       payload.State,
	}
	if err := s.consumeOAuthCodeAndCreateAccess(ctx, token.Id, expiredAt, access, refresh, accessPayload); err != nil {
		return nil, err
	}
	millis := expiredAt.UnixMilli()
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

func (s *serviceImpl) oauthAuthorizationCode(ctx context.Context, code string) (*entity.OauthToken, *oauthRuntimePayload, error) {
	return s.oauthTokenByCode(ctx, oauthCodePrefix+strings.TrimSpace(code), oauthKindAuthorizationCode)
}

func (s *serviceImpl) oauthAccessToken(ctx context.Context, accessToken string) (*oauthRuntimePayload, error) {
	var token *entity.OauthToken
	err := s.tenantFilter.Apply(ctx, dao.OauthToken.Ctx(ctx), "").
		Where(dao.OauthToken.Columns().Access, oauthAccessPrefix+strings.TrimSpace(accessToken)).
		Scan(&token)
	if err != nil {
		return nil, err
	}
	_, payload, err := parseOAuthRuntimeToken(token, oauthKindAccessToken)
	return payload, err
}

func (s *serviceImpl) oauthTokenByCode(ctx context.Context, code string, kind string) (*entity.OauthToken, *oauthRuntimePayload, error) {
	var token *entity.OauthToken
	err := s.tenantFilter.Apply(ctx, dao.OauthToken.Ctx(ctx), "").
		Where(dao.OauthToken.Columns().Code, code).
		Scan(&token)
	if err != nil {
		return nil, nil, err
	}
	return parseOAuthRuntimeToken(token, kind)
}

func (s *serviceImpl) consumeOAuthCodeAndCreateAccess(
	ctx context.Context,
	codeTokenID int64,
	expiredAt time.Time,
	access string,
	refresh string,
	payload oauthRuntimePayload,
) error {
	content, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	tenantID, actorID := s.baseOwnedDO(ctx, true)
	return dao.OauthToken.Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {
		result, err := tx.Model(dao.OauthToken.Table()).Safe().Ctx(ctx).
			Where(plugincontract.TenantFilterColumn, tenantID).
			Where(dao.OauthToken.Columns().Id, codeTokenID).
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
		if _, err := tx.Model(dao.OauthToken.Table()).Safe().Ctx(ctx).
			Data(do.OauthToken{
				TenantId:  tenantID,
				ExpiredAt: &expiredAt,
				Code:      oauthAccessPrefix + access,
				Access:    oauthAccessPrefix + access,
				Refresh:   oauthRefreshPrefix + refresh,
				Data:      string(content),
				CreatedBy: actorID,
				UpdatedBy: actorID,
			}).
			Insert(); err != nil {
			return err
		}
		_, err = tx.Model(dao.OauthLog.Table()).Safe().Ctx(ctx).
			Data(do.OauthLog{
				TenantId:    tenantID,
				UserId:      payload.AccountID,
				AppId:       payload.AppID,
				RedirectUri: payload.RedirectURI,
				Scope:       payload.Scope,
				CreatedBy:   actorID,
				UpdatedBy:   actorID,
			}).
			Insert()
		return err
	})
}

func parseOAuthRuntimeToken(token *entity.OauthToken, kind string) (*entity.OauthToken, *oauthRuntimePayload, error) {
	if token == nil || token.ExpiredAt == nil || token.ExpiredAt.Before(time.Now()) {
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
	return trimmed == "" || trimmed == oauthGrantTypeAuthorizationCode
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
