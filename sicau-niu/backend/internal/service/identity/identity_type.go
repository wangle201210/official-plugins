// identity_type.go defines the player identity-type enum as a Go named type with
// stable constants and validation. The identity type is a small, plugin-private,
// single-language enum (per the program design decision to keep it as named
// constants rather than a host dictionary), so it is governed here in code.

package identity

// IdentityType is the player identity tag enum. Its string value is persisted in
// the player table identity_type column.
type IdentityType string

const (
	// IdentityTypeStudent marks an enrolled student player.
	IdentityTypeStudent IdentityType = "student"
	// IdentityTypeAlumni marks an alumnus player.
	IdentityTypeAlumni IdentityType = "alumni"
	// IdentityTypeFriend marks a SICAU-friend (社会好友) player.
	IdentityTypeFriend IdentityType = "friend"
)

// String returns the persisted string form of the identity type.
func (t IdentityType) String() string {
	return string(t)
}

// valid reports whether t is one of the allowed identity types.
func (t IdentityType) valid() bool {
	switch t {
	case IdentityTypeStudent, IdentityTypeAlumni, IdentityTypeFriend:
		return true
	default:
		return false
	}
}

// requiresCollege reports whether the identity type requires a college and grade
// selection. Students must pick a college; alumni and friends may omit it.
func (t IdentityType) requiresCollege() bool {
	return t == IdentityTypeStudent
}
