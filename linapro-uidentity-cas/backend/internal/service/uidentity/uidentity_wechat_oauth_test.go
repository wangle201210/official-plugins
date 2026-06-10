// This file verifies the built-in Wechat OAuth authorize URL layout and code
// resolution against a local HTTP stub without external dependencies.

package uidentity

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

// TestBuiltinWechatAuthorizeURLDisabledWithoutCredentials keeps the external
// adapter contract: no credentials means no built-in URL.
func TestBuiltinWechatAuthorizeURLDisabledWithoutCredentials(t *testing.T) {
	t.Parallel()

	service := &serviceImpl{configSvc: newLegacyConfigTestService(t, "")}
	got, err := service.builtinWechatAuthorizeURL(context.Background(), configKeyLegacyWechatLoginCallbackURL, nil, "state-1")
	if err != nil {
		t.Fatalf("authorize URL failed: %v", err)
	}
	if got != "" {
		t.Fatalf("expected empty URL without credentials, got %q", got)
	}
}

// TestBuiltinWechatAuthorizeURLMatchesOldOAuthLayout verifies the official
// open.weixin.qq.com URL shape used by the old silenceper GetRedirectURL.
func TestBuiltinWechatAuthorizeURLMatchesOldOAuthLayout(t *testing.T) {
	t.Parallel()

	service := &serviceImpl{configSvc: newLegacyConfigTestService(t, `
legacy:
  wechat:
    appId: wx123
    appSecret: shh
    loginCallbackUrl: https://cas.example/api/v1/cas/loginByQr
`)}
	got, err := service.builtinWechatAuthorizeURL(context.Background(), configKeyLegacyWechatLoginCallbackURL, map[string]string{
		"appid":       "portal",
		"cascallback": "https://app.example/cb",
	}, "state-1")
	if err != nil {
		t.Fatalf("authorize URL failed: %v", err)
	}
	if !strings.HasPrefix(got, defaultLegacyWechatAuthorizeBase+"?") || !strings.HasSuffix(got, "#wechat_redirect") {
		t.Fatalf("authorize URL shape mismatch: %q", got)
	}
	parsed, err := url.Parse(strings.TrimSuffix(got, "#wechat_redirect"))
	if err != nil {
		t.Fatalf("parse authorize URL: %v", err)
	}
	query := parsed.Query()
	if query.Get("appid") != "wx123" || query.Get("response_type") != "code" ||
		query.Get("scope") != defaultLegacyWechatScope || query.Get("state") != "state-1" {
		t.Fatalf("authorize URL query mismatch: %q", got)
	}
	redirect, err := url.Parse(query.Get("redirect_uri"))
	if err != nil {
		t.Fatalf("parse redirect_uri: %v", err)
	}
	if redirect.Host != "cas.example" || redirect.Query().Get("appid") != "portal" ||
		redirect.Query().Get("cascallback") != "https://app.example/cb" {
		t.Fatalf("redirect_uri mismatch: %q", query.Get("redirect_uri"))
	}
}

// TestResolveBuiltinWechatUnionIDExchangesCode verifies the official
// access_token exchange, the snapshot-user guard, and error payload handling.
func TestResolveBuiltinWechatUnionIDExchangesCode(t *testing.T) {
	t.Parallel()

	responseBody := `{"openid":"o1","unionid":"u-trusted"}`
	var gotQuery url.Values
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/sns/oauth2/access_token" {
			t.Errorf("access_token path = %q", r.URL.Path)
		}
		gotQuery = r.URL.Query()
		_, _ = fmt.Fprint(w, responseBody)
	}))
	t.Cleanup(server.Close)

	service := &serviceImpl{configSvc: newLegacyConfigTestService(t, fmt.Sprintf(`
legacy:
  wechat:
    appId: wx123
    appSecret: shh
    apiBase: %s
`, server.URL))}
	unionID, err := service.resolveBuiltinWechatUnionID(context.Background(), "code-1")
	if err != nil {
		t.Fatalf("resolve union ID failed: %v", err)
	}
	if unionID != "u-trusted" {
		t.Fatalf("union ID = %q, want u-trusted", unionID)
	}
	if gotQuery.Get("appid") != "wx123" || gotQuery.Get("secret") != "shh" ||
		gotQuery.Get("code") != "code-1" || gotQuery.Get("grant_type") != "authorization_code" {
		t.Fatalf("access_token query mismatch: %#v", gotQuery)
	}

	responseBody = `{"openid":"o2","unionid":"u-snapshot","is_snapshotuser":1}`
	unionID, err = service.resolveBuiltinWechatUnionID(context.Background(), "code-2")
	if err != nil {
		t.Fatalf("resolve snapshot union ID failed: %v", err)
	}
	if unionID != "" {
		t.Fatalf("snapshot user union ID must be rejected, got %q", unionID)
	}

	responseBody = `{"errcode":40029,"errmsg":"invalid code"}`
	unionID, err = service.resolveBuiltinWechatUnionID(context.Background(), "code-3")
	if err != nil {
		t.Fatalf("resolve failed-code union ID errored: %v", err)
	}
	if unionID != "" {
		t.Fatalf("failed exchange must resolve empty union ID, got %q", unionID)
	}
}
