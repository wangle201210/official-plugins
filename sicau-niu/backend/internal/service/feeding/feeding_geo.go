// feeding_geo.go implements the great-circle distance used to decide the
// iron-cow proximity bonus. Plugin coordinates are GCJ-02 end to end, so the
// Haversine formula is applied directly on GCJ-02 points (the shared local
// datum offset cancels at activity scale); iron-cow positions synced from the
// IOT platform must be converted to GCJ-02 before they reach the iron table.
// The feeding capability keeps its own copy so it does not reach across the
// activation package-internal boundary.

package feeding

import "math"

// earthRadiusMeters is the mean Earth radius in meters used by the Haversine
// great-circle distance.
const earthRadiusMeters = 6371000.0

// haversineMeters returns the great-circle distance in meters between two
// latitude/longitude points in decimal degrees.
func haversineMeters(lat1, lng1, lat2, lng2 float64) float64 {
	phi1 := lat1 * math.Pi / 180
	phi2 := lat2 * math.Pi / 180
	deltaPhi := (lat2 - lat1) * math.Pi / 180
	deltaLambda := (lng2 - lng1) * math.Pi / 180

	a := math.Sin(deltaPhi/2)*math.Sin(deltaPhi/2) +
		math.Cos(phi1)*math.Cos(phi2)*math.Sin(deltaLambda/2)*math.Sin(deltaLambda/2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
	return earthRadiusMeters * c
}
