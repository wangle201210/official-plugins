// This file implements the built-in Wechat OAuth integration that the old
// uidentity/admin embedded through the silenceper official-account SDK:
// building open.weixin.qq.com authorize URLs and resolving OAuth codes to
// union IDs through api.weixin.qq.com. It activates when plugin config
// provides a Wechat appId/appSecret; the external adapter URLs configured via
// runtime.wechat* keys still take precedence so existing deployments keep
// their delegated flows.

package uidentity

import (
	"context"
	"encoding/json"
	"net/url"
	"strings"
	"time"

	"github.com/gogf/gf/v2/frame/g"

	"lina-core/pkg/logger"
)

const (
	configKeyLegacyWechatAppID         = "legacy.wechat.appId"
	configKeyLegacyWechatAppSecret     = "legacy.wechat.appSecret"
	configKeyLegacyWechatScope         = "legacy.wechat.scope"
	configKeyLegacyWechatAuthorizeBase = "legacy.wechat.authorizeBase"
	configKeyLegacyWechatAPIBase       = "legacy.wechat.apiBase"

	// Public callback addresses of this deployment's own legacy endpoints,
	// embedded as redirect_uri in the Wechat authorize URL. They mirror the
	// old wechat.loginRedirect / activeRedirect / bindRedirect settings.
	configKeyLegacyWechatLoginCallbackURL      = "legacy.wechat.loginCallbackUrl"
	configKeyLegacyWechatActivationCallbackURL = "legacy.wechat.activationCallbackUrl"
	configKeyLegacyWechatRebindCallbackURL     = "legacy.wechat.rebindCallbackUrl"

	// Defaults match the old silenceper official-account OAuth endpoints.
	defaultLegacyWechatScope         = "snsapi_userinfo"
	defaultLegacyWechatAuthorizeBase = "https://open.weixin.qq.com/connect/oauth2/authorize"
	defaultLegacyWechatAPIBase      = "https://api.weixin.qq.com"
)

// legacyWechatOAuthConfig carries one resolved built-in Wechat OAuth setup.
type legacyWechatOAuthConfig struct {
	AppID         string
	AppSecret     string
	Scope         string
	AuthorizeBase string
	APIBase       string
}

// enabled reports whether built-in Wechat OAuth can run.
func (c *legacyWechatOAuthConfig) enabled() bool {
	return c != nil && strings.TrimSpace(c.AppID) != "" && strings.TrimSpace(c.AppSecret) != ""
}

// legacyWechatOAuthConfig reads built-in Wechat OAuth settings from config.
func (s *serviceImpl) legacyWechatOAuthConfig(ctx context.Context) (*legacyWechatOAuthConfig, error) {
	appID, err := s.configSvc.String(ctx, configKeyLegacyWechatAppID, "")
	if err != nil {
		return nil, err
	}
	appSecret, err := s.configSvc.String(ctx, configKeyLegacyWechatAppSecret, "")
	if err != nil {
		return nil, err
	}
	scope, err := s.configSvc.String(ctx, configKeyLegacyWechatScope, defaultLegacyWechatScope)
	if err != nil {
		return nil, err
	}
	authorizeBase, err := s.configSvc.String(ctx, configKeyLegacyWechatAuthorizeBase, defaultLegacyWechatAuthorizeBase)
	if err != nil {
		return nil, err
	}
	apiBase, err := s.configSvc.String(ctx, configKeyLegacyWechatAPIBase, defaultLegacyWechatAPIBase)
	if err != nil {
		return nil, err
	}
	return &legacyWechatOAuthConfig{
		AppID:         appID,
		AppSecret:     appSecret,
		Scope:         scope,
		AuthorizeBase: authorizeBase,
		APIBase:       apiBase,
	}, nil
}

// builtinWechatAuthorizeURL returns one official Wechat OAuth authorize URL
// for a plugin callback endpoint, matching the old GetRedirectURL layout.
func (s *serviceImpl) builtinWechatAuthorizeURL(ctx context.Context, callbackConfigKey string, callbackQuery map[string]string, state string) (string, error) {
	cfg, err := s.legacyWechatOAuthConfig(ctx)
	if err != nil {
		return "", err
	}
	if !cfg.enabled() {
		return "", nil
	}
	redirectURI, err := s.configSvc.String(ctx, callbackConfigKey, "")
	if err != nil {
		return "", err
	}
	if strings.TrimSpace(redirectURI) == "" {
		return "", nil
	}
	for key, value := range callbackQuery {
		if strings.TrimSpace(value) != "" {
			redirectURI = callbackWithQuery(redirectURI, key, value)
		}
	}
	return legacyWechatAuthorizeURL(cfg, redirectURI, state), nil
}

// legacyWechatAuthorizeURL builds the old open.weixin.qq.com authorize URL.
func legacyWechatAuthorizeURL(cfg *legacyWechatOAuthConfig, redirectURI string, state string) string {
	query := url.Values{}
	query.Set("appid", strings.TrimSpace(cfg.AppID))
	query.Set("redirect_uri", redirectURI)
	query.Set("response_type", "code")
	query.Set("scope", strings.TrimSpace(cfg.Scope))
	query.Set("state", strings.TrimSpace(state))
	return strings.TrimRight(cfg.AuthorizeBase, "/") + "?" + query.Encode() + "#wechat_redirect"
}

// legacyWechatAccessTokenResponse carries the official sns/oauth2/access_token result.
type legacyWechatAccessTokenResponse struct {
	OpenID         string `json:"openid"`
	UnionID        string `json:"unionid"`
	IsSnapshotUser int    `json:"is_snapshotuser"`
	ErrCode        int    `json:"errcode"`
	ErrMsg         string `json:"errmsg"`
}

// resolveBuiltinWechatUnionID exchanges one OAuth code through the official
// Wechat API and returns the trusted union ID. Snapshot (virtual) users and
// missing union IDs resolve to an empty string, matching the old guard that
// refused to bind untrusted identities.
func (s *serviceImpl) resolveBuiltinWechatUnionID(ctx context.Context, code string) (string, error) {
	cfg, err := s.legacyWechatOAuthConfig(ctx)
	if err != nil {
		return "", err
	}
	if !cfg.enabled() || strings.TrimSpace(code) == "" {
		return "", nil
	}
	response, err := g.Client().
		SetTimeout(10*time.Second).
		Get(ctx, strings.TrimRight(cfg.APIBase, "/")+"/sns/oauth2/access_token", map[string]string{
			"appid":      strings.TrimSpace(cfg.AppID),
			"secret":     strings.TrimSpace(cfg.AppSecret),
			"code":       strings.TrimSpace(code),
			"grant_type": "authorization_code",
		})
	if err != nil {
		return "", err
	}
	defer response.Close()
	payload := &legacyWechatAccessTokenResponse{}
	if err := json.Unmarshal(response.ReadAll(), payload); err != nil {
		return "", err
	}
	if payload.ErrCode != 0 {
		logger.Warningf(ctx, "legacy wechat access_token failed errcode=%d errmsg=%s", payload.ErrCode, payload.ErrMsg)
		return "", nil
	}
	if payload.IsSnapshotUser == 1 {
		// Old rule: snapshot users carry untrusted union IDs and must not bind.
		logger.Warningf(ctx, "legacy wechat snapshot user rejected openid=%s", payload.OpenID)
		return "", nil
	}
	return strings.TrimSpace(payload.UnionID), nil
}

// resolveLegacyWechatUnionID resolves one OAuth code to a union ID using the
// external adapter URL first and the built-in Wechat API as fallback.
func (s *serviceImpl) resolveLegacyWechatUnionID(ctx context.Context, code string) (string, error) {
	resolveURL, err := s.configSvc.String(ctx, configKeyWechatLoginCodeResolveURL, "")
	if err != nil {
		return "", err
	}
	if strings.TrimSpace(resolveURL) != "" {
		resp, err := g.Client().SetTimeout(10*time.Second).Get(ctx, resolveURL, map[string]any{"code": strings.TrimSpace(code)})
		if err != nil {
			return "", err
		}
		defer resp.Close()
		return wechatUnionIDFromPayload(resp.ReadAll()), nil
	}
	return s.resolveBuiltinWechatUnionID(ctx, code)
}
