// This file tests legacy CAS XML rendering helpers without database fixtures.

package uidentity

import (
	"strings"
	"testing"
)

func TestBuildLegacyCASSuccessXML(t *testing.T) {
	t.Parallel()

	out, err := buildLegacyCASSuccessXML(&RuntimeAccount{
		ID:          10,
		Number:      "A001",
		Name:        "Alice",
		Phone:       "13800000000",
		Status:      AccountStatusNormal,
		ContainerID: legacyCASStaffContainerID,
		UnitID:      20,
		Detail: &RuntimeAccountDetail{
			Birthday: "2000-01-01",
			Email:    "alice@example.com",
			Gender:   1,
			Idcard:   "510000200001010000",
		},
	})
	if err != nil {
		t.Fatalf("build legacy CAS success XML: %v", err)
	}
	xmlText := string(out.XML)
	for _, expected := range []string{
		`<cas:serviceResponse xmlns:cas="http://www.yale.edu/tp/cas">`,
		"<cas:authenticationSuccess>",
		"<cas:user>A001</cas:user>",
		"<cas:workcode>A001</cas:workcode>",
		"<cas:departmentid>20</cas:departmentid>",
		"<cas:email>alice@example.com</cas:email>",
		"<cas:userType>01</cas:userType>",
	} {
		if !strings.Contains(xmlText, expected) {
			t.Fatalf("expected XML to contain %q, got %s", expected, xmlText)
		}
	}
}

func TestBuildLegacyCASFailureXML(t *testing.T) {
	t.Parallel()

	out, err := buildLegacyCASFailureXML("ticket expired")
	if err != nil {
		t.Fatalf("build legacy CAS failure XML: %v", err)
	}
	xmlText := string(out.XML)
	for _, expected := range []string{
		"<cas:authenticationFailure code=\"INVALID_TICKET\">",
		"CAS ticket validation failed: ticket expired",
	} {
		if !strings.Contains(xmlText, expected) {
			t.Fatalf("expected XML to contain %q, got %s", expected, xmlText)
		}
	}
}
