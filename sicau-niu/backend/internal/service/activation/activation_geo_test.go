// activation_geo_test.go covers the Haversine great-circle distance helper used
// by the LBS activation gate. These are pure-logic tests with no database
// dependency.

package activation

import (
	"math"
	"testing"
)

// TestHaversineMetersZeroDistance verifies identical coordinates yield zero.
func TestHaversineMetersZeroDistance(t *testing.T) {
	got := haversineMeters(30.123456, 103.123456, 30.123456, 103.123456)
	if got != 0 {
		t.Fatalf("expected zero distance for identical points, got %f", got)
	}
}

// TestHaversineMetersKnownDistance verifies a known short offset stays within a
// tolerance of the analytic great-circle distance. One degree of latitude is
// about 111.2 km; a 0.001 degree step is roughly 111 meters.
func TestHaversineMetersKnownDistance(t *testing.T) {
	got := haversineMeters(30.0, 103.0, 30.001, 103.0)
	const wantApprox = 111.0
	if math.Abs(got-wantApprox) > 2.0 {
		t.Fatalf("expected ~%.1f meters, got %f", wantApprox, got)
	}
}

// TestHaversineMetersThresholdBoundary verifies a point just inside and just
// outside a 50-meter threshold is classified correctly, exercising the LBS gate
// comparison the activation flow performs.
func TestHaversineMetersThresholdBoundary(t *testing.T) {
	const threshold = 50.0
	// ~0.0003 degree latitude offset is about 33 meters: inside the threshold.
	inside := haversineMeters(30.0, 103.0, 30.0003, 103.0)
	if inside > threshold {
		t.Fatalf("expected inside-threshold distance, got %f", inside)
	}
	// ~0.001 degree latitude offset is about 111 meters: outside the threshold.
	outside := haversineMeters(30.0, 103.0, 30.001, 103.0)
	if outside <= threshold {
		t.Fatalf("expected outside-threshold distance, got %f", outside)
	}
}
