// card_type_test.go verifies the card category enum validation and persisted
// string form. These are pure-logic, same-package tests that need no database.

package card

import "testing"

// TestCategoryValid verifies the allowed card categories pass and any other value
// is rejected.
func TestCategoryValid(t *testing.T) {
	valid := []Category{
		CategoryPerson,
		CategoryEvent,
		CategoryResearch,
		CategoryCollege,
		CategorySpirit,
	}
	for _, c := range valid {
		if !c.valid() {
			t.Fatalf("expected category %q to be valid", c)
		}
	}

	invalid := []Category{"", "PERSON", "person ", "place", "object"}
	for _, c := range invalid {
		if c.valid() {
			t.Fatalf("expected category %q to be invalid", c)
		}
	}
}

// TestCategoryString verifies the persisted string form matches the constant
// value.
func TestCategoryString(t *testing.T) {
	cases := map[Category]string{
		CategoryPerson:   "person",
		CategoryEvent:    "event",
		CategoryResearch: "research",
		CategoryCollege:  "college",
		CategorySpirit:   "spirit",
	}
	for c, want := range cases {
		if got := c.String(); got != want {
			t.Fatalf("expected %q, got %q", want, got)
		}
	}
}

// TestNormalizeEnabled verifies any non-zero enabled input clamps to 1 and zero
// stays 0 so the persisted flag is always 0 or 1.
func TestNormalizeEnabled(t *testing.T) {
	cases := map[int]int{0: 0, 1: 1, 2: 1, -3: 1, 100: 1}
	for in, want := range cases {
		if got := normalizeEnabled(in); got != want {
			t.Fatalf("normalizeEnabled(%d) = %d, want %d", in, got, want)
		}
	}
}

// TestValidateQuoteContentRejectsBlank verifies a nil input or a blank/whitespace
// content is rejected with CodeQuoteContentRequired, and a padded content is
// trimmed. This is pure logic and runs even when the DB harness skips.
func TestValidateQuoteContentRejectsBlank(t *testing.T) {
	for _, in := range []*QuoteMutateInput{nil, {Content: ""}, {Content: "   "}} {
		_, err := validateQuoteContent(in)
		assertBizCode(t, err, CodeQuoteContentRequired.RuntimeCode())
	}

	content, err := validateQuoteContent(&QuoteMutateInput{Content: "  Strive on  "})
	if err != nil {
		t.Fatalf("expected valid content, got %v", err)
	}
	if content != "Strive on" {
		t.Fatalf("expected trimmed content %q, got %q", "Strive on", content)
	}
}
