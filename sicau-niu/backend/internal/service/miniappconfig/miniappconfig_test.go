// miniappconfig_test.go verifies public defaults, live rule projection,
// timezone handling and stable fuzzy-area behavior.

package miniappconfig

import (
	"math"
	"testing"
	"time"

	rulessvc "lina-plugin-sicau-niu/backend/internal/service/rules"
)

func TestSnapshotContainsAllCampusCalibrations(t *testing.T) {
	svc := New(nil, Config{DefaultCampus: CampusChengdu, Anniversary: "120 周年校庆", AnniversaryAt: "2026-10-06", ActivateRadiusMeters: 60})
	out := cloneSnapshot(&svc.(*serviceImpl).defaults)
	if len(out.Campuses) != 3 || out.DefaultCampus != CampusChengdu || out.ActivateRadiusM != 60 || out.Debug {
		t.Fatalf("unexpected public config: %+v", out)
	}
	if out.StealDailyLimit != 5 || out.GiftDailyLimit != 12 || out.GiftMinAmount != 12 {
		t.Fatalf("unexpected public gameplay defaults: %+v", out)
	}
	for _, campus := range out.Campuses {
		if campus.Calibration.ImageSize.W <= 0 || campus.Calibration.ImageSize.H <= 0 || campus.Calibration.Affine.A == 0 || campus.Calibration.Affine.D == 0 {
			t.Fatalf("campus calibration is incomplete: %+v", campus)
		}
	}
}

func TestSnapshotProjectsPublicGameplayRules(t *testing.T) {
	rules := &rulessvc.RuleSet{
		ActivationLBSThresholdMeters: 75,
		StealDailyLimit:              7,
		GiftDailyLimit:               9,
		GiftMinAmount:                18,
	}
	svc := New(nil, Config{ActivateRadiusMeters: 60}).(*serviceImpl)
	out := cloneSnapshot(&svc.defaults)

	applyRuntimeRules(out, rules)
	if out.ActivateRadiusM != 75 || out.StealDailyLimit != 7 || out.GiftDailyLimit != 9 || out.GiftMinAmount != 18 {
		t.Fatalf("unexpected public gameplay rules: out=%+v", out)
	}
}

func TestCountdownDaysUsesMiniProgramTimezone(t *testing.T) {
	now := time.Date(2026, time.October, 5, 16, 30, 0, 0, time.UTC)
	if got := countdownDays("2026-10-06", now); got != 0 {
		t.Fatalf("countdownDays() = %d, want 0 after local midnight", got)
	}
}

func TestFuzzyAreaIsStableAndContainsTrueAnchor(t *testing.T) {
	const lat, lng = 30.7058, 103.8318
	first := FuzzyArea(17, lat, lng)
	second := FuzzyArea(17, lat, lng)
	if *first != *second {
		t.Fatalf("fuzzy area changed between requests: first=%+v second=%+v", first, second)
	}
	if first.Lat == lat && first.Lng == lng {
		t.Fatal("fuzzy center must not equal the true anchor")
	}
	meters := math.Hypot((first.Lat-lat)*110540, (first.Lng-lng)*111320*math.Cos(lat*math.Pi/180))
	if meters >= float64(first.RadiusM) {
		t.Fatalf("true anchor fell outside fuzzy area: offset=%f radius=%d", meters, first.RadiusM)
	}
}
