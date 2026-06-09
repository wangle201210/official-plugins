// cattle_type_test.go verifies the cattle stable enums: type/subtype/status
// validation, persisted string form and the subtype/college requirement helpers.
// These are pure-logic, same-package tests that need no database.

package cattle

import "testing"

// TestNiuTypeValid verifies the allowed cattle types pass and any other value is
// rejected.
func TestNiuTypeValid(t *testing.T) {
	valid := []NiuType{NiuTypeCommon, NiuTypeSpecial}
	for _, nt := range valid {
		if !nt.valid() {
			t.Fatalf("expected cattle type %q to be valid", nt)
		}
	}

	invalid := []NiuType{"", "COMMON", "special ", "rare", "iron"}
	for _, nt := range invalid {
		if nt.valid() {
			t.Fatalf("expected cattle type %q to be invalid", nt)
		}
	}
}

// TestNiuTypeString verifies the persisted string form matches the constant
// value.
func TestNiuTypeString(t *testing.T) {
	cases := map[NiuType]string{
		NiuTypeCommon:  "common",
		NiuTypeSpecial: "special",
	}
	for nt, want := range cases {
		if got := nt.String(); got != want {
			t.Fatalf("expected %q, got %q", want, got)
		}
	}
}

// TestNiuTypeRequiresSubtype verifies only special cattle require a subtype and
// name; common cattle do not.
func TestNiuTypeRequiresSubtype(t *testing.T) {
	if !NiuTypeSpecial.requiresSubtype() {
		t.Fatal("expected special cattle to require a subtype")
	}
	if NiuTypeCommon.requiresSubtype() {
		t.Fatal("expected common cattle not to require a subtype")
	}
}

// TestSpecialSubtypeValid verifies the allowed special subtypes pass and any
// other value is rejected. The empty subtype is invalid because special cattle
// always carry one.
func TestSpecialSubtypeValid(t *testing.T) {
	valid := []SpecialSubtype{
		SpecialSubtypeCollege,
		SpecialSubtypeContribution,
		SpecialSubtypeAlumni,
		SpecialSubtypeSpirit,
	}
	for _, st := range valid {
		if !st.valid() {
			t.Fatalf("expected special subtype %q to be valid", st)
		}
	}

	invalid := []SpecialSubtype{"", "COLLEGE", "college ", "teacher", "campus"}
	for _, st := range invalid {
		if st.valid() {
			t.Fatalf("expected special subtype %q to be invalid", st)
		}
	}
}

// TestSpecialSubtypeString verifies the persisted string form matches the
// constant value.
func TestSpecialSubtypeString(t *testing.T) {
	cases := map[SpecialSubtype]string{
		SpecialSubtypeCollege:      "college",
		SpecialSubtypeContribution: "contribution",
		SpecialSubtypeAlumni:       "alumni",
		SpecialSubtypeSpirit:       "spirit",
	}
	for st, want := range cases {
		if got := st.String(); got != want {
			t.Fatalf("expected %q, got %q", want, got)
		}
	}
}

// TestSpecialSubtypeRequiresCollege verifies only the college subtype requires a
// linked college.
func TestSpecialSubtypeRequiresCollege(t *testing.T) {
	if !SpecialSubtypeCollege.requiresCollege() {
		t.Fatal("expected college subtype to require a linked college")
	}
	for _, st := range []SpecialSubtype{SpecialSubtypeContribution, SpecialSubtypeAlumni, SpecialSubtypeSpirit} {
		if st.requiresCollege() {
			t.Fatalf("expected subtype %q not to require a linked college", st)
		}
	}
}

// TestNiuStatusString verifies the persisted string form of the cattle status
// matches the constant value.
func TestNiuStatusString(t *testing.T) {
	cases := map[NiuStatus]string{
		NiuStatusInactive: "inactive",
		NiuStatusActive:   "active",
	}
	for st, want := range cases {
		if got := st.String(); got != want {
			t.Fatalf("expected %q, got %q", want, got)
		}
	}
}
