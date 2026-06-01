// This file tests OAuth runtime helpers that preserve legacy client behavior.

package uidentity

import "testing"

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
	if oauthGrantTypeSupported("password") {
		t.Fatalf("expected password grant type to be rejected")
	}
}
