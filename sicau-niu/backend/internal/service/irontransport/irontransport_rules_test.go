package irontransport

import (
	"math"
	"testing"
)

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
