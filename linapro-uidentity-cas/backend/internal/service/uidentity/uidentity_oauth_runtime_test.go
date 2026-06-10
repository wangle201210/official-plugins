// This file tests OAuth runtime helpers that preserve legacy client behavior.

package uidentity

import (
	"testing"
	"time"
)

func TestOAuthClientSecretMatchesLegacyEscapedSecret(t *testing.T) {
	if !oauthClientSecretMatches("a+b secret", "a%2Bb+secret") {
		t.Fatalf("expected URL-escaped legacy client secret to match")
	}
	if oauthClientSecretMatches("a+b secret", "wrong") {
		t.Fatalf("expected wrong client secret to be rejected")
	}
}

func TestOAuthResolveRedirectURI(t *testing.T) {
	got, err := oauthResolveRedirectURI("https://example.com/oauth/callback", "")
	if err != nil {
		t.Fatalf("expected default callback redirect to pass: %v", err)
	}
	if got != "https://example.com/oauth/callback" {
		t.Fatalf("unexpected redirect URI: %s", got)
	}
	if _, err := oauthResolveRedirectURI("https://example.com/oauth/callback", "https://evil.example.com/oauth/callback"); err == nil {
		t.Fatalf("expected mismatched redirect URI to be rejected")
	}
}

func TestOAuthGrantTypeSupported(t *testing.T) {
	if !oauthGrantTypeSupported("") {
		t.Fatalf("expected empty grant type to default to authorization_code")
	}
	if !oauthGrantTypeSupported(oauthGrantTypeAuthorizationCode) {
		t.Fatalf("expected authorization_code grant type to be supported")
	}
	if !oauthGrantTypeSupported(oauthGrantTypeRefreshToken) {
		t.Fatalf("expected refresh_token grant type to be supported")
	}
	if oauthGrantTypeSupported("password") {
		t.Fatalf("expected password grant type to be rejected")
	}
}

func TestOAuthAccessPayloadExpired(t *testing.T) {
	now := time.Now()
	if oauthAccessPayloadExpired(nil, now) {
		t.Fatalf("expected nil payload to never report access expiry")
	}
	legacyRow := &oauthRuntimePayload{}
	if oauthAccessPayloadExpired(legacyRow, now) {
		t.Fatalf("expected rows without AccessExpiredAt to rely on row expiry only")
	}
	live := &oauthRuntimePayload{AccessExpiredAt: now.Add(time.Hour).UnixMilli()}
	if oauthAccessPayloadExpired(live, now) {
		t.Fatalf("expected future access expiry to stay valid")
	}
	expired := &oauthRuntimePayload{AccessExpiredAt: now.Add(-time.Second).UnixMilli()}
	if !oauthAccessPayloadExpired(expired, now) {
		t.Fatalf("expected past access expiry to be rejected")
	}
}
