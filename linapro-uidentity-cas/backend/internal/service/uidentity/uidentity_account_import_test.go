// This file verifies pure account import parsing helpers without depending on
// runtime database fixtures.

package uidentity

import "testing"

// TestImportRowBlank verifies legacy import skips effectively empty rows.
func TestImportRowBlank(t *testing.T) {
	t.Parallel()

	if !importRowBlank([]string{"", " "}) {
		t.Fatal("expected blank row to be skipped")
	}
	if importRowBlank([]string{"A001", "Alice"}) {
		t.Fatal("expected non-empty row to be imported")
	}
}

// TestImportBirthdayNormalizesDate verifies birthday values remain date-only.
func TestImportBirthdayNormalizesDate(t *testing.T) {
	t.Parallel()

	if got := importBirthday("2026-06-01 08:30:00"); got != "2026-06-01" {
		t.Fatalf("expected date-only birthday, got %q", got)
	}
}

// TestValidSMSType verifies supported legacy SMS scenarios.
func TestValidSMSType(t *testing.T) {
	t.Parallel()

	for _, smsType := range []string{smsTypeCasLogin, smsTypeCasActive, smsTypeCasBind, smsTypePasswordReset} {
		if !validSMSType(smsType) {
			t.Fatalf("expected %s to be valid", smsType)
		}
	}
	if validSMSType("unknown") {
		t.Fatal("expected unknown SMS type to be rejected")
	}
}
