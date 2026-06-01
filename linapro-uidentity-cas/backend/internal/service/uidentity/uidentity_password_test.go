// This file verifies pure password helper behavior without depending on the
// runtime PostgreSQL fixture used by DAO-backed service methods.

package uidentity

import (
	"strings"
	"testing"

	"lina-plugin-linapro-uidentity-cas/backend/internal/model/entity"
)

// TestHashPasswordIsStable verifies password hashes are deterministic and not
// returned as plaintext.
func TestHashPasswordIsStable(t *testing.T) {
	t.Parallel()

	first := hashPassword("S3cure@2026")
	second := hashPassword("S3cure@2026")
	if first != second {
		t.Fatal("expected password hash to be stable")
	}
	if first == "S3cure@2026" {
		t.Fatal("expected password hash to differ from plaintext")
	}
	if len(first) != 64 {
		t.Fatalf("expected SHA-256 hex length 64, got %d", len(first))
	}
}

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

// TestPasswordMatchesHashedAccount verifies runtime login compares the stored
// hash instead of accepting plaintext or empty account data.
func TestPasswordMatchesHashedAccount(t *testing.T) {
	t.Parallel()

	account := &entity.Account{PasswordHash: hashPassword("S3cure@2026")}
	if !passwordMatches(account, "S3cure@2026") {
		t.Fatal("expected hashed password to match")
	}
	if passwordMatches(account, "wrong") {
		t.Fatal("expected wrong password to fail")
	}
	if passwordMatches(nil, "S3cure@2026") {
		t.Fatal("expected nil account to fail")
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
