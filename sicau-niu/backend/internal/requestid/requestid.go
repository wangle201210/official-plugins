// Package requestid normalizes player-scoped idempotency keys shared by write
// services. The public API and PostgreSQL varchar columns define the limit in
// Unicode characters, so validation must not use UTF-8 byte length.
package requestid

import (
	"strings"
	"unicode/utf8"
)

const maxCharacters = 64

// Normalize trims surrounding whitespace and validates the shared character
// limit. It returns the normalized value only when the key is usable.
func Normalize(value string) (string, bool) {
	value = strings.TrimSpace(value)
	return value, value != "" && utf8.RuneCountInString(value) <= maxCharacters
}
