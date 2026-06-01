// This file verifies pure password helper behavior without depending on the
// runtime PostgreSQL fixture used by DAO-backed service methods.

package uidentity

import (
	"strings"
	"testing"
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
