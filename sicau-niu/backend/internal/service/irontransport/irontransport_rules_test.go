// irontransport_rules_test.go covers bounded cloud-moving input validation and
// distance edge cases without database dependencies. It verifies request IDs
// use the public character-count contract rather than UTF-8 byte length.
package irontransport

import (
	"math"
	"strings"
	"testing"
)

func TestNormalizeRequestIDCountsUnicodeCharacters(t *testing.T) {
	if value, ok := normalizeRequestID(strings.Repeat("牛", 64)); !ok || value == "" {
		t.Fatal("expected 64 Unicode characters to be accepted")
	}
	if _, ok := normalizeRequestID(strings.Repeat("牛", 65)); ok {
		t.Fatal("expected 65 Unicode characters to be rejected")
	}
}

func TestValidCoordinateRejectsNonFiniteValues(t *testing.T) {
	for _, coordinates := range [][2]float64{
		{math.NaN(), 0},
		{0, math.NaN()},
		{math.Inf(1), 0},
		{0, math.Inf(-1)},
	} {
		if validCoordinate(coordinates[0], coordinates[1]) {
			t.Fatalf("expected non-finite coordinate to be rejected: %#v", coordinates)
		}
	}
}

func TestHaversineMetersHandlesAntipodalPoints(t *testing.T) {
	distance := haversineMeters(0, 0, 0, 180)
	if distance < 20_015_000 || distance > 20_016_000 {
		t.Fatalf("unexpected antipodal distance: %d", distance)
	}
}
