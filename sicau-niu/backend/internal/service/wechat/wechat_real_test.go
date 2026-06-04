// wechat_real_test.go unit-tests the pure WeChat response parsers and error
// mapping. These tests exercise the production gateway's parsing logic without any
// network access or real credentials; the real WeChat round-trip is an external
// dependency that cannot run in CI.

package wechat

import (
	"testing"

	"lina-core/pkg/bizerr"
)

// assertCode fails the test unless err is a structured business error whose runtime
// code equals wantCode.
func assertCode(t *testing.T, err error, wantCode string) {
	t.Helper()
	if err == nil {
		t.Fatalf("expected error with code %s, got nil", wantCode)
	}
	bizErr, ok := bizerr.As(err)
	if !ok {
		t.Fatalf("expected structured business error with code %s, got %T: %v", wantCode, err, err)
	}
	if bizErr.RuntimeCode() != wantCode {
		t.Fatalf("expected code %s, got %s", wantCode, bizErr.RuntimeCode())
	}
}

func TestParseSessionResponse(t *testing.T) {
	openid, err := parseSessionResponse(`{"openid":"oABC123","session_key":"sk","unionid":"u"}`)
	if err != nil {
		t.Fatalf("success parse returned error: %v", err)
	}
	if openid != "oABC123" {
		t.Fatalf("openid = %q, want oABC123", openid)
	}

	_, err = parseSessionResponse(`{"errcode":40029,"errmsg":"invalid code"}`)
	assertCode(t, err, "PLUGIN_SICAU_NIU_WECHAT_CODE_INVALID")

	_, err = parseSessionResponse(`{"session_key":"sk"}`)
	assertCode(t, err, "PLUGIN_SICAU_NIU_WECHAT_CODE_INVALID")

	_, err = parseSessionResponse(`not-json`)
	assertCode(t, err, "PLUGIN_SICAU_NIU_WECHAT_CODE_INVALID")
}

func TestParseTokenResponse(t *testing.T) {
	token, ttl, err := parseTokenResponse(`{"access_token":"TOK","expires_in":7200}`)
	if err != nil {
		t.Fatalf("success parse returned error: %v", err)
	}
	if token != "TOK" || ttl != 7200 {
		t.Fatalf("token/ttl = %q/%d, want TOK/7200", token, ttl)
	}

	// A tiny TTL is bumped above the refresh buffer so the cache window stays positive.
	_, ttl, err = parseTokenResponse(`{"access_token":"T","expires_in":10}`)
	if err != nil {
		t.Fatalf("tiny-ttl parse returned error: %v", err)
	}
	if ttl <= tokenRefreshBufferSeconds {
		t.Fatalf("tiny ttl = %d, want > %d", ttl, tokenRefreshBufferSeconds)
	}

	_, _, err = parseTokenResponse(`{"errcode":40013,"errmsg":"invalid appid"}`)
	assertCode(t, err, "PLUGIN_SICAU_NIU_WECHAT_REQUEST_FAILED")

	_, _, err = parseTokenResponse(`{}`)
	assertCode(t, err, "PLUGIN_SICAU_NIU_WECHAT_REQUEST_FAILED")
}

func TestParsePhoneResponse(t *testing.T) {
	phone, err := parsePhoneResponse(`{"errcode":0,"errmsg":"ok","phone_info":{"phoneNumber":"13800138000","purePhoneNumber":"13800138000"}}`)
	if err != nil {
		t.Fatalf("success parse returned error: %v", err)
	}
	if phone != "13800138000" {
		t.Fatalf("phone = %q, want 13800138000", phone)
	}

	_, err = parsePhoneResponse(`{"errcode":40029,"errmsg":"invalid code"}`)
	assertCode(t, err, "PLUGIN_SICAU_NIU_WECHAT_PHONE_DECODE_FAILED")

	_, err = parsePhoneResponse(`{"errcode":0,"phone_info":{"phoneNumber":""}}`)
	assertCode(t, err, "PLUGIN_SICAU_NIU_WECHAT_PHONE_DECODE_FAILED")
}
