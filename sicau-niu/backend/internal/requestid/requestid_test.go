// requestid_test.go verifies the shared idempotency-key contract uses Unicode
// character count, trims surrounding whitespace and rejects empty or overlong
// values without any database dependency.
package requestid

import (
	"strings"
	"testing"
)

func TestNormalizeUsesUnicodeCharacterCount(t *testing.T) {
	value, ok := Normalize("  " + strings.Repeat("牛", 64) + "  ")
	if !ok || value != strings.Repeat("牛", 64) {
		t.Fatalf("expected 64 Unicode characters after trim, got %q ok=%v", value, ok)
	}
	if _, ok = Normalize(strings.Repeat("牛", 65)); ok {
		t.Fatal("expected 65 Unicode characters to be rejected")
	}
	if _, ok = Normalize("   "); ok {
		t.Fatal("expected whitespace-only key to be rejected")
	}
}
