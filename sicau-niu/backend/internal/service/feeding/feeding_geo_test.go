// feeding_geo_test.go covers the Haversine distance used for the iron-bonus
// decision. These are pure-logic checks that always run without a database.

package feeding

import (
	"math"
	"testing"
)

// TestHaversineSamePoint verifies two identical coordinates are zero meters apart.
func TestHaversineSamePoint(t *testing.T) {
	if d := haversineMeters(30.0, 103.0, 30.0, 103.0); d != 0 {
		t.Fatalf("expected 0 meters for the same point, got %f", d)
	}
}

// TestHaversineWithinThreshold verifies a sub-12m offset is reported as a small
// distance under the bonus threshold.
func TestHaversineWithinThreshold(t *testing.T) {
	// ~0.00005 degrees latitude is roughly 5.6m, within the 12m bonus threshold.
	d := haversineMeters(30.0, 103.0, 30.00005, 103.0)
	if d <= 0 || d >= 12 {
		t.Fatalf("expected a small distance under 12m, got %f", d)
	}
}

// TestHaversineKnownDistance verifies a 0.01-degree latitude offset is about
// 1.11km, matching the great-circle expectation within tolerance.
func TestHaversineKnownDistance(t *testing.T) {
	d := haversineMeters(30.0, 103.0, 30.01, 103.0)
	const want = 1111.95
	if math.Abs(d-want) > 5 {
		t.Fatalf("expected ~%.2fm for 0.01 degree latitude, got %f", want, d)
	}
}
