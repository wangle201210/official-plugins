// card_code_test.go verifies that the card and quote business error codes expose
// stable runtime codes and derived i18n message keys.

package card

import (
	"testing"

	"lina-core/pkg/bizerr"
)

// TestCardBusinessErrorMetadata verifies a representative subset of card and quote
// codes expose stable runtime codes and derived message keys so the HTTP/i18n
// contract does not drift.
func TestCardBusinessErrorMetadata(t *testing.T) {
	cases := []struct {
		code       *bizerr.Code
		runtime    string
		messageKey string
	}{
		{
			code:       CodeCardNiuTaken,
			runtime:    "PLUGIN_SICAU_NIU_CARD_NIU_TAKEN",
			messageKey: "error.plugin.sicau.niu.card.niu.taken",
		},
		{
			code:       CodeCardNiuInvalid,
			runtime:    "PLUGIN_SICAU_NIU_CARD_NIU_INVALID",
			messageKey: "error.plugin.sicau.niu.card.niu.invalid",
		},
		{
			code:       CodeQuoteContentRequired,
			runtime:    "PLUGIN_SICAU_NIU_QUOTE_CONTENT_REQUIRED",
			messageKey: "error.plugin.sicau.niu.quote.content.required",
		},
	}

	for _, tc := range cases {
		err := bizerr.NewCode(tc.code)
		bizErr, ok := bizerr.As(err)
		if !ok {
			t.Fatalf("expected structured business error, got %T", err)
		}
		if bizErr.RuntimeCode() != tc.runtime {
			t.Fatalf("expected runtime code %s, got %s", tc.runtime, bizErr.RuntimeCode())
		}
		if bizErr.MessageKey() != tc.messageKey {
			t.Fatalf("expected message key %s, got %s", tc.messageKey, bizErr.MessageKey())
		}
	}
}
