// cattle_code_test.go verifies that the cattle and iron-cow business error codes
// expose stable runtime codes and derived i18n message keys.

package cattle

import (
	"testing"

	"lina-core/pkg/bizerr"
)

// TestCattleBusinessErrorMetadata verifies a representative subset of cattle and
// iron-cow codes expose stable runtime codes and derived message keys so the
// HTTP/i18n contract does not drift.
func TestCattleBusinessErrorMetadata(t *testing.T) {
	cases := []struct {
		code       *bizerr.Code
		runtime    string
		messageKey string
	}{
		{
			code:       CodeNiuCodeExists,
			runtime:    "PLUGIN_SICAU_NIU_NIU_CODE_EXISTS",
			messageKey: "error.plugin.sicau.niu.niu.code.exists",
		},
		{
			code:       CodeNiuCollegeInvalid,
			runtime:    "PLUGIN_SICAU_NIU_NIU_COLLEGE_INVALID",
			messageKey: "error.plugin.sicau.niu.niu.college.invalid",
		},
		{
			code:       CodeIronCodeExists,
			runtime:    "PLUGIN_SICAU_NIU_IRON_CODE_EXISTS",
			messageKey: "error.plugin.sicau.niu.iron.code.exists",
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
