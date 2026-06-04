// wechat_real.go holds the production WeChat gateway. It performs the real WeChat
// open-API HTTP flows — `sns/jscode2session` for login and the new-style
// `getPhoneNumber` code flow (`cgi-bin/token` + `wxa/business/getuserphonenumber`)
// for phone authorization — and maps WeChat error codes onto the gateway business
// errors. The only external dependency is the configured AppID/Secret: once real
// credentials are configured and `wechat.mock=false`, this gateway runs for real
// with no further code change. The `access_token` is cached in-process behind a
// mutex and refreshed ahead of its expiry so phone decoding does not re-mint a
// token on every call and trip WeChat's global token rate limit. Response parsing
// is isolated into pure functions so the mapping logic is unit-tested without
// network access or credentials; the HTTP calls are thin wrappers around those.
package wechat

import (
	"context"
	"fmt"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/gogf/gf/v2/encoding/gjson"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"

	"lina-core/pkg/bizerr"
)

// WeChat open-API endpoints and the access-token refresh buffer.
const (
	wechatAPIBase = "https://api.weixin.qq.com"
	sessionPath   = "/sns/jscode2session"
	tokenPath     = "/cgi-bin/token"
	phonePath     = "/wxa/business/getuserphonenumber"
	// tokenRefreshBufferSeconds refreshes the cached access_token this many seconds
	// before its reported expiry so an in-flight call never uses a just-expired token.
	tokenRefreshBufferSeconds = 300
)

// realGateway is the production WeChat gateway. It holds the credentials for the
// real open-API calls plus an in-process access_token cache guarded by a mutex so
// the gateway is safe for concurrent use.
type realGateway struct {
	appID  string // appID is the WeChat mini-program AppID.
	secret string // secret is the WeChat mini-program AppSecret.

	mu          sync.Mutex // mu guards the cached access_token and its expiry.
	cachedToken string     // cachedToken is the last fetched access_token.
	tokenExpiry time.Time  // tokenExpiry is when cachedToken must be refreshed.
}

// newRealGateway creates the production WeChat gateway.
func newRealGateway(appID, secret string) Gateway {
	return &realGateway{appID: appID, secret: secret}
}

// Code2Session exchanges a WeChat login code for the player's openid via the real
// `sns/jscode2session` call.
func (r *realGateway) Code2Session(ctx context.Context, code string) (string, error) {
	if strings.TrimSpace(code) == "" {
		return "", bizerr.NewCode(CodeWeChatCodeInvalid)
	}
	endpoint := fmt.Sprintf(
		"%s%s?appid=%s&secret=%s&js_code=%s&grant_type=authorization_code",
		wechatAPIBase, sessionPath,
		url.QueryEscape(r.appID), url.QueryEscape(r.secret), url.QueryEscape(code),
	)
	body, err := httpGet(ctx, endpoint)
	if err != nil {
		return "", bizerr.WrapCode(err, CodeWeChatRequestFailed)
	}
	return parseSessionResponse(body)
}

// DecodePhone decodes the phone number via the new-style getPhoneNumber code flow:
// it fetches (and caches) an access_token, then exchanges the phone code through
// `wxa/business/getuserphonenumber`.
func (r *realGateway) DecodePhone(ctx context.Context, in DecodePhoneInput) (string, error) {
	code := strings.TrimSpace(in.Code)
	if code == "" {
		// The real gateway implements the recommended new-style code flow; the
		// legacy encryptedData+IV flow needs the login session_key and is not wired.
		return "", bizerr.NewCode(CodeWeChatPhoneDecodeFailed)
	}
	token, err := r.accessToken(ctx)
	if err != nil {
		return "", err
	}
	endpoint := fmt.Sprintf("%s%s?access_token=%s", wechatAPIBase, phonePath, url.QueryEscape(token))
	body, err := httpPostJSON(ctx, endpoint, g.Map{"code": code})
	if err != nil {
		return "", bizerr.WrapCode(err, CodeWeChatRequestFailed)
	}
	return parsePhoneResponse(body)
}

// accessToken returns a valid WeChat access_token, reusing the cached token while
// it is fresh and refreshing it (under the mutex) once it nears expiry.
func (r *realGateway) accessToken(ctx context.Context) (string, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.cachedToken != "" && time.Now().Before(r.tokenExpiry) {
		return r.cachedToken, nil
	}
	endpoint := fmt.Sprintf(
		"%s%s?grant_type=client_credential&appid=%s&secret=%s",
		wechatAPIBase, tokenPath, url.QueryEscape(r.appID), url.QueryEscape(r.secret),
	)
	body, err := httpGet(ctx, endpoint)
	if err != nil {
		return "", bizerr.WrapCode(err, CodeWeChatRequestFailed)
	}
	token, ttlSeconds, err := parseTokenResponse(body)
	if err != nil {
		return "", err
	}
	r.cachedToken = token
	r.tokenExpiry = time.Now().Add(time.Duration(ttlSeconds-tokenRefreshBufferSeconds) * time.Second)
	return token, nil
}

// httpGet performs a GET and returns the response body string.
func httpGet(ctx context.Context, endpoint string) (string, error) {
	resp, err := g.Client().Get(ctx, endpoint)
	if err != nil {
		return "", err
	}
	defer func() { _ = resp.Close() }()
	return resp.ReadAllString(), nil
}

// httpPostJSON performs a JSON POST and returns the response body string.
func httpPostJSON(ctx context.Context, endpoint string, data interface{}) (string, error) {
	resp, err := g.Client().ContentJson().Post(ctx, endpoint, data)
	if err != nil {
		return "", err
	}
	defer func() { _ = resp.Close() }()
	return resp.ReadAllString(), nil
}

// parseSessionResponse extracts the openid from a `jscode2session` response and
// maps a non-zero errcode or empty openid to CodeWeChatCodeInvalid.
func parseSessionResponse(body string) (string, error) {
	j, err := gjson.DecodeToJson([]byte(body))
	if err != nil {
		return "", bizerr.WrapCode(err, CodeWeChatCodeInvalid)
	}
	if errcode := j.Get("errcode").Int(); errcode != 0 {
		return "", bizerr.WrapCode(
			gerror.Newf("wechat jscode2session errcode %d: %s", errcode, j.Get("errmsg").String()),
			CodeWeChatCodeInvalid,
		)
	}
	openid := strings.TrimSpace(j.Get("openid").String())
	if openid == "" {
		return "", bizerr.NewCode(CodeWeChatCodeInvalid)
	}
	return openid, nil
}

// parseTokenResponse extracts the access_token and its expiry from a `cgi-bin/token`
// response and maps a non-zero errcode or empty token to a request failure.
func parseTokenResponse(body string) (string, int, error) {
	j, err := gjson.DecodeToJson([]byte(body))
	if err != nil {
		return "", 0, bizerr.WrapCode(err, CodeWeChatRequestFailed)
	}
	if errcode := j.Get("errcode").Int(); errcode != 0 {
		return "", 0, bizerr.WrapCode(
			gerror.Newf("wechat token errcode %d: %s", errcode, j.Get("errmsg").String()),
			CodeWeChatRequestFailed,
		)
	}
	token := strings.TrimSpace(j.Get("access_token").String())
	if token == "" {
		return "", 0, bizerr.NewCode(CodeWeChatRequestFailed)
	}
	ttl := j.Get("expires_in").Int()
	if ttl <= tokenRefreshBufferSeconds {
		// Guard against a tiny or missing TTL so the refresh buffer never yields a
		// non-positive cache window.
		ttl = tokenRefreshBufferSeconds + 60
	}
	return token, ttl, nil
}

// parsePhoneResponse extracts the phone number from a `getuserphonenumber` response
// and maps a non-zero errcode or empty phone to CodeWeChatPhoneDecodeFailed.
func parsePhoneResponse(body string) (string, error) {
	j, err := gjson.DecodeToJson([]byte(body))
	if err != nil {
		return "", bizerr.WrapCode(err, CodeWeChatPhoneDecodeFailed)
	}
	if errcode := j.Get("errcode").Int(); errcode != 0 {
		return "", bizerr.WrapCode(
			gerror.Newf("wechat getuserphonenumber errcode %d: %s", errcode, j.Get("errmsg").String()),
			CodeWeChatPhoneDecodeFailed,
		)
	}
	phone := strings.TrimSpace(j.Get("phone_info.phoneNumber").String())
	if phone == "" {
		return "", bizerr.NewCode(CodeWeChatPhoneDecodeFailed)
	}
	return phone, nil
}
