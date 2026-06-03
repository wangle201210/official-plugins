// wechat_mock_test.go verifies the mockable WeChat gateway used in C1: fixed and
// code-derived openids from Code2Session, and phone resolution from the override
// or code in DecodePhone. All assertions are pure-logic and need no database.

package wechat

import (
	"context"
	"strings"
	"testing"

	"lina-core/pkg/bizerr"
)

// TestMockCode2SessionReturnsFixedOpenid verifies a configured mock openid is
// returned for every login code.
func TestMockCode2SessionReturnsFixedOpenid(t *testing.T) {
	ctx := context.Background()
	gateway := New(Config{Mock: true, MockOpenid: "fixed-openid"})

	for _, code := range []string{"code-1", "code-2"} {
		openid, err := gateway.Code2Session(ctx, code)
		if err != nil {
			t.Fatalf("code2session failed for %q: %v", code, err)
		}
		if openid != "fixed-openid" {
			t.Fatalf("expected fixed openid, got %q", openid)
		}
	}
}

// TestMockCode2SessionDerivesStableOpenid verifies that with no fixed openid the
// gateway derives a stable openid from the code: the same code yields the same
// openid and different codes yield different openids.
func TestMockCode2SessionDerivesStableOpenid(t *testing.T) {
	ctx := context.Background()
	gateway := New(Config{Mock: true})

	first, err := gateway.Code2Session(ctx, "stable-code")
	if err != nil {
		t.Fatalf("code2session failed: %v", err)
	}
	again, err := gateway.Code2Session(ctx, "stable-code")
	if err != nil {
		t.Fatalf("code2session failed: %v", err)
	}
	if first != again {
		t.Fatalf("expected stable openid for same code, got %q then %q", first, again)
	}
	if !strings.HasPrefix(first, mockOpenidPrefix) {
		t.Fatalf("expected derived openid prefix %q, got %q", mockOpenidPrefix, first)
	}

	other, err := gateway.Code2Session(ctx, "different-code")
	if err != nil {
		t.Fatalf("code2session failed: %v", err)
	}
	if other == first {
		t.Fatal("expected different codes to derive different openids")
	}
}

// TestMockCode2SessionRejectsEmptyCode verifies a blank code is rejected with
// CodeWeChatCodeInvalid.
func TestMockCode2SessionRejectsEmptyCode(t *testing.T) {
	ctx := context.Background()
	gateway := New(Config{Mock: true})

	_, err := gateway.Code2Session(ctx, "   ")
	if err == nil {
		t.Fatal("expected blank code to be rejected")
	}
	bizErr, ok := bizerr.As(err)
	if !ok {
		t.Fatalf("expected structured business error, got %T", err)
	}
	if bizErr.RuntimeCode() != CodeWeChatCodeInvalid.RuntimeCode() {
		t.Fatalf("expected %s, got %s", CodeWeChatCodeInvalid.RuntimeCode(), bizErr.RuntimeCode())
	}
}

// TestMockDecodePhonePrefersOverride verifies DecodePhone returns the explicit
// override when set, otherwise falls back to the code, and rejects an empty
// payload with CodeWeChatPhoneDecodeFailed.
func TestMockDecodePhonePrefersOverride(t *testing.T) {
	ctx := context.Background()
	gateway := New(Config{Mock: true})

	phone, err := gateway.DecodePhone(ctx, DecodePhoneInput{PhoneOverride: "13900000000", Code: "ignored"})
	if err != nil {
		t.Fatalf("decode phone failed: %v", err)
	}
	if phone != "13900000000" {
		t.Fatalf("expected override phone, got %q", phone)
	}

	phone, err = gateway.DecodePhone(ctx, DecodePhoneInput{Code: "13911111111"})
	if err != nil {
		t.Fatalf("decode phone from code failed: %v", err)
	}
	if phone != "13911111111" {
		t.Fatalf("expected code-derived phone, got %q", phone)
	}

	_, err = gateway.DecodePhone(ctx, DecodePhoneInput{})
	if err == nil {
		t.Fatal("expected empty phone payload to be rejected")
	}
	bizErr, ok := bizerr.As(err)
	if !ok {
		t.Fatalf("expected structured business error, got %T", err)
	}
	if bizErr.RuntimeCode() != CodeWeChatPhoneDecodeFailed.RuntimeCode() {
		t.Fatalf("expected %s, got %s", CodeWeChatPhoneDecodeFailed.RuntimeCode(), bizErr.RuntimeCode())
	}
}
