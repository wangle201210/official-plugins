// identity_code_test.go verifies that the player identity business error codes
// expose stable runtime codes and derived i18n message keys.

package identity

import (
	"testing"

	"lina-core/pkg/bizerr"
)

// TestIdentityBusinessErrorMetadata verifies a representative subset of identity
// codes expose stable runtime codes and derived message keys.
func TestIdentityBusinessErrorMetadata(t *testing.T) {
	cases := []struct {
		code       *bizerr.Code
		runtime    string
		messageKey string
	}{
		{
			code:       CodePhoneTaken,
			runtime:    "PLUGIN_SICAU_NIU_PHONE_TAKEN",
			messageKey: "error.plugin.sicau.niu.phone.taken",
		},
		{
			code:       CodeCollegeInvalid,
			runtime:    "PLUGIN_SICAU_NIU_IDENTITY_COLLEGE_INVALID",
			messageKey: "error.plugin.sicau.niu.identity.college.invalid",
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
