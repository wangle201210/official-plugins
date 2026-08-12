// irontransport_rules.go validates bounded inputs and computes contribution distance.
package irontransport

import (
	"math"
	"strings"
	"time"
)

func normalizeRequestID(value string) (string, bool) {
	value = strings.TrimSpace(value)
	return value, value != "" && len(value) <= 64
}

func normalizeName(value string) (string, bool) {
	value = strings.TrimSpace(value)
	return value, value != "" && len([]rune(value)) <= 64
}

func normalizePagination(in *PageInput) (int, int) {
	if in == nil {
		return defaultPageNum, defaultPageSize
	}
	pageNum := in.PageNum
	if pageNum <= 0 {
		pageNum = defaultPageNum
	}
	pageSize := in.PageSize
	if pageSize <= 0 {
		pageSize = defaultPageSize
	}
	if pageSize > maxPageSize {
		pageSize = maxPageSize
	}
	return pageNum, pageSize
}

func validCoordinate(lat, lng float64) bool {
	return !math.IsNaN(lat) && !math.IsInf(lat, 0) &&
		!math.IsNaN(lng) && !math.IsInf(lng, 0) &&
		lat >= -90 && lat <= 90 && lng >= -180 && lng <= 180
}

func haversineMeters(lat1, lng1, lat2, lng2 float64) int64 {
	const earthRadius = 6_371_000.0
	phi1 := lat1 * math.Pi / 180
	phi2 := lat2 * math.Pi / 180
	deltaPhi := (lat2 - lat1) * math.Pi / 180
	deltaLambda := (lng2 - lng1) * math.Pi / 180
	a := math.Sin(deltaPhi/2)*math.Sin(deltaPhi/2) + math.Cos(phi1)*math.Cos(phi2)*math.Sin(deltaLambda/2)*math.Sin(deltaLambda/2)
	a = math.Max(0, math.Min(1, a))
	return int64(math.Round(earthRadius * 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))))
}

func validActivityDate(value string) bool {
	if strings.TrimSpace(value) == "" {
		return true
	}
	_, err := time.Parse("2006-01-02", value)
	return err == nil
}
