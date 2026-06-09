// cattle_type.go defines the cattle stable enums as Go named types with
// constants and validation: cattle type (common/special), special subtype
// (college/contribution/alumni/spirit) and cattle lifecycle status. These are
// small, plugin-private, single-language enums governed in code per the program
// design decision to keep them as named constants rather than host dictionary
// entries.

package cattle

// NiuType is the cattle type enum. Its string value is persisted in the niu
// table niu_type column.
type NiuType string

const (
	// NiuTypeCommon marks a common cattle without name/subtype/college.
	NiuTypeCommon NiuType = "common"
	// NiuTypeSpecial marks a special cattle that carries a name and subtype.
	NiuTypeSpecial NiuType = "special"
)

// String returns the persisted string form of the cattle type.
func (t NiuType) String() string {
	return string(t)
}

// valid reports whether t is one of the allowed cattle types.
func (t NiuType) valid() bool {
	switch t {
	case NiuTypeCommon, NiuTypeSpecial:
		return true
	default:
		return false
	}
}

// requiresSubtype reports whether the cattle type requires a special subtype and
// name. Only special cattle carry a subtype; common cattle must omit it.
func (t NiuType) requiresSubtype() bool {
	return t == NiuTypeSpecial
}

// SpecialSubtype is the special-cattle subtype enum. Its string value is
// persisted in the niu table special_subtype column; it is empty for common
// cattle.
type SpecialSubtype string

const (
	// SpecialSubtypeCollege marks a college special cattle linked to a college.
	SpecialSubtypeCollege SpecialSubtype = "college"
	// SpecialSubtypeContribution marks a contribution special cattle.
	SpecialSubtypeContribution SpecialSubtype = "contribution"
	// SpecialSubtypeAlumni marks an alumni special cattle.
	SpecialSubtypeAlumni SpecialSubtype = "alumni"
	// SpecialSubtypeSpirit marks a spirit special cattle.
	SpecialSubtypeSpirit SpecialSubtype = "spirit"
)

// String returns the persisted string form of the special subtype.
func (t SpecialSubtype) String() string {
	return string(t)
}

// valid reports whether t is one of the allowed special subtypes.
func (t SpecialSubtype) valid() bool {
	switch t {
	case SpecialSubtypeCollege, SpecialSubtypeContribution, SpecialSubtypeAlumni, SpecialSubtypeSpirit:
		return true
	default:
		return false
	}
}

// requiresCollege reports whether the subtype requires a linked college. Only
// college special cattle link to a college from the C1 dictionary.
func (t SpecialSubtype) requiresCollege() bool {
	return t == SpecialSubtypeCollege
}

// NiuStatus is the cattle lifecycle-status enum. Its string value is persisted
// in the niu table status column; the activation flow (C3) manages transitions.
type NiuStatus string

const (
	// NiuStatusInactive marks a not-yet-activated cattle; it is the create default.
	NiuStatusInactive NiuStatus = "inactive"
	// NiuStatusActive marks an activated cattle.
	NiuStatusActive NiuStatus = "active"
)

// String returns the persisted string form of the cattle status.
func (s NiuStatus) String() string {
	return string(s)
}
