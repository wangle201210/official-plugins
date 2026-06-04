// activation_geo.go implements the LBS distance check used to gate activation. It
// computes the great-circle distance between two WGS-84 coordinates with the
// Haversine formula using only the standard library; there is no image
// recognition, the photo is evidence only.

package activation

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
