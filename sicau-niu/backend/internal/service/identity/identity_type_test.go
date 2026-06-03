// identity_type_test.go verifies the player identity-type enum validation and
// the college-requirement helper. These are pure-logic, same-package tests that
// need no database.

package identity

import "testing"

// TestIdentityTypeValid verifies the allowed identity types pass and any other
// value is rejected.
func TestIdentityTypeValid(t *testing.T) {
	valid := []IdentityType{IdentityTypeStudent, IdentityTypeAlumni, IdentityTypeFriend}
	for _, it := range valid {
		if !it.valid() {
			t.Fatalf("expected identity type %q to be valid", it)
		}
	}

	invalid := []IdentityType{"", "teacher", "STUDENT", "student ", "guest"}
	for _, it := range invalid {
		if it.valid() {
			t.Fatalf("expected identity type %q to be invalid", it)
		}
	}
}

// TestIdentityTypeString verifies the persisted string form matches the constant
// value.
func TestIdentityTypeString(t *testing.T) {
	cases := map[IdentityType]string{
		IdentityTypeStudent: "student",
		IdentityTypeAlumni:  "alumni",
		IdentityTypeFriend:  "friend",
	}
	for it, want := range cases {
		if got := it.String(); got != want {
			t.Fatalf("expected %q, got %q", want, got)
		}
	}
}

// TestIdentityTypeRequiresCollege verifies only students require a college and
// grade selection.
func TestIdentityTypeRequiresCollege(t *testing.T) {
	if !IdentityTypeStudent.requiresCollege() {
		t.Fatal("expected student identity to require a college")
	}
	if IdentityTypeAlumni.requiresCollege() {
		t.Fatal("expected alumni identity not to require a college")
	}
	if IdentityTypeFriend.requiresCollege() {
		t.Fatal("expected friend identity not to require a college")
	}
}
