// card_type.go defines the card category stable enum as a Go named type with
// constants and validation. The category is a small, plugin-private,
// single-language enum governed in code per the program design decision to keep
// it as named constants rather than a host dictionary entry.

package card

// Category is the card category enum. Its string value is persisted in the card
// table category column.
type Category string

const (
	// CategoryPerson marks a person card.
	CategoryPerson Category = "person"
	// CategoryEvent marks an event card.
	CategoryEvent Category = "event"
	// CategoryResearch marks a research card.
	CategoryResearch Category = "research"
	// CategoryCollege marks a college card.
	CategoryCollege Category = "college"
	// CategorySpirit marks a spirit card.
	CategorySpirit Category = "spirit"
)

// String returns the persisted string form of the card category.
func (c Category) String() string {
	return string(c)
}

// valid reports whether c is one of the allowed card categories.
func (c Category) valid() bool {
	switch c {
	case CategoryPerson, CategoryEvent, CategoryResearch, CategoryCollege, CategorySpirit:
		return true
	default:
		return false
	}
}

// ValidCategory reports whether value is one of the allowed card categories. It
// is the exported contract used by the C3 collection filter to validate an
// optional category against the same enum the card content asset persists,
// without duplicating the category set across packages.
func ValidCategory(value string) bool {
	return Category(value).valid()
}
