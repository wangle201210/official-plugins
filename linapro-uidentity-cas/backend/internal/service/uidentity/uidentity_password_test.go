// This file verifies pure password helper behavior without depending on the
// runtime PostgreSQL fixture used by DAO-backed service methods.

package uidentity

import (
	"strings"
	"testing"
)

// TestRandomTokenUsesPrefix verifies generated runtime tokens carry the
// caller-provided domain prefix.
func TestRandomTokenUsesPrefix(t *testing.T) {
	t.Parallel()

	token, err := randomToken("code")
	if err != nil {
		t.Fatalf("generate token: %v", err)
	}
	if !strings.HasPrefix(token, "code_") {
		t.Fatalf("expected token prefix code_, got %q", token)
	}
	if len(token) <= len("code_") {
		t.Fatal("expected token to include random payload")
	}
}

// TestCallbackWithTicketPreservesExistingQuery verifies CAS callback URL
// decoration keeps existing query parameters and appends the service ticket.
func TestCallbackWithTicketPreservesExistingQuery(t *testing.T) {
	t.Parallel()

	got := callbackWithTicket("https://example.com/callback?locale=zh-CN", "ST_123")
	if got != "https://example.com/callback?locale=zh-CN&ticket=ST_123" {
		t.Fatalf("unexpected callback URL: %s", got)
	}
}

func TestPasswordFailureCodeUsesLegacyDomain(t *testing.T) {
	t.Parallel()

	if got := passwordFailureCode(" A001 "); got != "cas:pwd:errnum:A001" {
		t.Fatalf("unexpected password failure code: %s", got)
	}
}

func TestPasswordFailureCodesFiltersBlankNumbers(t *testing.T) {
	t.Parallel()

	got := passwordFailureCodes([]string{" A001 ", "", "B002"})
	if len(got) != 2 || got[0] != "cas:pwd:errnum:A001" || got[1] != "cas:pwd:errnum:B002" {
		t.Fatalf("unexpected password failure codes: %#v", got)
	}
}

func TestValidateLegacyLDAPPasswordMatchesOldFormats(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name     string
		password string
		ldapPass string
		want     bool
	}{
		{name: "plain", password: "secret", ldapPass: "secret", want: true},
		{name: "plain mismatch", password: "secret", ldapPass: "other", want: false},
		{name: "ssha", password: "123456", ldapPass: "{SSHA}dx7V6wZKPRnO6MlJHxpqKebdDKg+1mti8RaZMw==", want: true},
		{name: "ssha mismatch", password: "12345", ldapPass: "{SSHA}q1k6zow91JW4Y/9n8N/kGj9A1J+u6BbJF9xFGQ==", want: false},
		{name: "md5", password: "123456", ldapPass: "{MD5}4QrcOUm6Wau+VuBX8g+IPg==", want: true},
		{name: "md5 mismatch", password: "12345", ldapPass: "{MD5}4QrcOUm6Wau+VuBX8g+IPg==", want: false},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			if got := validateLegacyLDAPPassword(tc.password, tc.ldapPass); got != tc.want {
				t.Fatalf("validateLegacyLDAPPassword() = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestGenerateLegacyLDAPSSHAValidates(t *testing.T) {
	t.Parallel()

	ldapPass, err := generateLegacyLDAPSSHA("changed-password")
	if err != nil {
		t.Fatalf("generateLegacyLDAPSSHA: %v", err)
	}
	if !strings.HasPrefix(ldapPass, ldapPasswordPrefixSSHA) {
		t.Fatalf("expected SSHA prefix, got %q", ldapPass)
	}
	if !validateLegacyLDAPPassword("changed-password", ldapPass) {
		t.Fatal("expected generated SSHA password to validate")
	}
}

func TestUniqueNonEmptyStringsTrimsAndCaps(t *testing.T) {
	t.Parallel()

	got := uniqueNonEmptyStrings([]string{" A001 ", "A001", "", "B002", "C003"}, 2)
	if len(got) != 2 || got[0] != "A001" || got[1] != "B002" {
		t.Fatalf("unexpected unique numbers: %#v", got)
	}
}
