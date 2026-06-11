// activation_activate_test.go covers the DB-gated LBS activation flow: GPS
// check-in matching, visible/inactive candidate gating, concurrent
// first-activator uniqueness, the per-day limit and on-activation card issuance.
// Each test is self-contained: it truncates the plugin tables, stages its own
// cattle/cards and asserts only on its own rows.

package activation

import (
	"context"
	"strconv"
	"sync"
	"testing"
	"time"

	"lina-plugin-sicau-niu/backend/internal/dao"
	"lina-plugin-sicau-niu/backend/internal/model/do"
	entitymodel "lina-plugin-sicau-niu/backend/internal/model/entity"
	cattlesvc "lina-plugin-sicau-niu/backend/internal/service/cattle"
)

// newActivationServiceForTest builds an activation service with a stub identity
// service and the basic poster renderer for DB-gated tests. The LBS threshold is
// fixed at 50 meters.
func newActivationServiceForTest() *serviceImpl {
	return &serviceImpl{
		identitySvc:    &fakeIdentityService{},
		posterRenderer: NewBasicPosterRenderer(),
		lbsThreshold:   50,
		campusBadge:    "TEST-BADGE",
	}
}

// stageActivatableNiu inserts an inactive cattle at a fixed anchor and returns its
// ID. The anchor is the origin used by the in-range/out-of-range assertions.
func stageActivatableNiu(t *testing.T, ctx context.Context) int64 {
	t.Helper()
	past := time.Now().Add(-time.Hour)
	return insertNiuRow(t, ctx, do.Niu{
		Code:     "NIU-ACT-001",
		NiuType:  cattlesvc.NiuTypeCommon.String(),
		Lat:      30.0,
		Lng:      103.0,
		OnlineAt: &past,
		Status:   cattlesvc.NiuStatusInactive.String(),
	})
}

func activationAttemptsForPlayer(t *testing.T, ctx context.Context, playerID int64) []*entitymodel.ActivationAttempt {
	t.Helper()
	rows := make([]*entitymodel.ActivationAttempt, 0)
	err := dao.ActivationAttempt.Ctx(ctx).
		Where(dao.ActivationAttempt.Columns().UserId, playerID).
		OrderAsc(dao.ActivationAttempt.Columns().Id).
		Scan(&rows)
	if err != nil {
		t.Fatalf("query activation attempts failed: %v", err)
	}
	return rows
}

// TestActivateNoNearbyNiuRejected verifies a location beyond the LBS threshold is
// rejected with CodeNoNearbyNiu and no activation row is written.
func TestActivateNoNearbyNiuRejected(t *testing.T) {
	ctx := context.Background()
	setupPostgreSQLActivationDB(t, ctx)
	svc := newActivationServiceForTest()
	niuID := stageActivatableNiu(t, ctx)
	playerID := insertUserRow(t, ctx, do.User{Openid: "openid-far"})

	// ~0.01 degree latitude offset is well over a kilometer, far outside 50m.
	_, err := svc.Activate(ctx, playerID, &ActivateInput{Lat: 30.01, Lng: 103.0})
	assertBizCode(t, err, CodeNoNearbyNiu.RuntimeCode())

	count, countErr := dao.Activation.Ctx(ctx).Where(dao.Activation.Columns().NiuId, niuID).Count()
	if countErr != nil {
		t.Fatalf("count activations failed: %v", countErr)
	}
	if count != 0 {
		t.Fatalf("expected no activation rows, got %d", count)
	}
	attempts := activationAttemptsForPlayer(t, ctx, playerID)
	if len(attempts) != 1 {
		t.Fatalf("expected one failed attempt row, got %d", len(attempts))
	}
	attempt := attempts[0]
	if attempt.Result != string(activationAttemptOutOfRange) {
		t.Fatalf("expected out-of-range attempt, got %q", attempt.Result)
	}
	if attempt.NearestNiuId != niuID || attempt.NiuId != 0 || attempt.DistanceM <= 50 || attempt.ThresholdM != 50 {
		t.Fatalf("attempt audit fields not recorded correctly: %+v", attempt)
	}
}

// TestActivateInvisibleNiuRejected verifies a direct activation call cannot
// bypass any player-visible-list gate and no activation row is written.
func TestActivateInvisibleNiuRejected(t *testing.T) {
	ctx := context.Background()
	setupPostgreSQLActivationDB(t, ctx)
	svc := newActivationServiceForTest()

	now := time.Now()
	past := now.Add(-time.Hour)
	future := now.Add(time.Hour)
	excludedWeekday := isoWeekday(now)%7 + 1

	cases := []struct {
		name string
		row  do.Niu
	}{
		{
			name: "online time empty",
			row: do.Niu{
				Code:    "NIU-ACT-NO-ONLINE",
				NiuType: cattlesvc.NiuTypeCommon.String(),
				Lat:     30.0,
				Lng:     103.0,
				Status:  cattlesvc.NiuStatusInactive.String(),
			},
		},
		{
			name: "online time in future",
			row: do.Niu{
				Code:     "NIU-ACT-FUTURE",
				NiuType:  cattlesvc.NiuTypeCommon.String(),
				Lat:      30.0,
				Lng:      103.0,
				OnlineAt: &future,
				Status:   cattlesvc.NiuStatusInactive.String(),
			},
		},
		{
			name: "weekday window excludes now",
			row: do.Niu{
				Code:            "NIU-ACT-WEEKDAY",
				NiuType:         cattlesvc.NiuTypeCommon.String(),
				Lat:             30.0,
				Lng:             103.0,
				OnlineAt:        &past,
				VisibleWeekdays: strconv.Itoa(excludedWeekday),
				Status:          cattlesvc.NiuStatusInactive.String(),
			},
		},
		{
			name: "time window excludes now",
			row: do.Niu{
				Code:         "NIU-ACT-TIME",
				NiuType:      cattlesvc.NiuTypeCommon.String(),
				Lat:          30.0,
				Lng:          103.0,
				OnlineAt:     &past,
				VisibleStart: oneMinuteAfter(now),
				VisibleEnd:   oneMinuteAfter(now),
				Status:       cattlesvc.NiuStatusInactive.String(),
			},
		},
	}

	for i, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			niuID := insertNiuRow(t, ctx, tc.row)
			playerID := insertUserRow(t, ctx, do.User{Openid: "openid-invisible-" + strconv.Itoa(i)})

			_, err := svc.Activate(ctx, playerID, &ActivateInput{Lat: 30.0, Lng: 103.0})
			assertBizCode(t, err, CodeNoNearbyNiu.RuntimeCode())

			count, countErr := dao.Activation.Ctx(ctx).Where(dao.Activation.Columns().NiuId, niuID).Count()
			if countErr != nil {
				t.Fatalf("count activations failed: %v", countErr)
			}
			if count != 0 {
				t.Fatalf("expected no activation rows, got %d", count)
			}
		})
	}
}

// oneMinuteAfter returns an HH:MM clock window bound that is guaranteed to be
// after now's current minute, wrapping at midnight.
func oneMinuteAfter(now time.Time) string {
	minute := (now.Hour()*60 + now.Minute() + 1) % (24 * 60)
	return time.Date(2000, 1, 1, minute/60, minute%60, 0, 0, time.Local).Format("15:04")
}

// TestActivateFirstActivatorFlipsStatus verifies the first in-range activation
// becomes the first-activator with order 1, flips the cattle to active and issues
// the cattle main card.
func TestActivateFirstActivatorFlipsStatus(t *testing.T) {
	ctx := context.Background()
	setupPostgreSQLActivationDB(t, ctx)
	svc := newActivationServiceForTest()
	niuID := stageActivatableNiu(t, ctx)
	insertCardRow(t, ctx, do.Card{
		NiuId:    niuID,
		Category: "spirit",
		Title:    "川农大精神",
		Content:  "爱国敬业",
	})
	playerID := insertUserRow(t, ctx, do.User{Openid: "openid-first"})

	out, err := svc.Activate(ctx, playerID, &ActivateInput{Lat: 30.0, Lng: 103.0})
	if err != nil {
		t.Fatalf("first activation failed: %v", err)
	}
	if out.NiuId != niuID {
		t.Fatalf("expected matched niu %d, got %d", niuID, out.NiuId)
	}
	if !out.IsFirst || out.OrderNo != 1 {
		t.Fatalf("expected first activator with order 1, got isFirst=%v order=%d", out.IsFirst, out.OrderNo)
	}
	if out.Card == nil || out.Card.Title != "川农大精神" {
		t.Fatalf("expected issued main card, got %+v", out.Card)
	}
	attempts := activationAttemptsForPlayer(t, ctx, playerID)
	if len(attempts) != 1 {
		t.Fatalf("expected one success attempt row, got %d", len(attempts))
	}
	attempt := attempts[0]
	if attempt.Result != string(activationAttemptSuccess) || attempt.NiuId != niuID || attempt.NearestNiuId != niuID {
		t.Fatalf("success attempt audit fields not recorded correctly: %+v", attempt)
	}

	statusVar, statusErr := dao.Niu.Ctx(ctx).
		Where(do.Niu{Id: niuID}).
		Fields(dao.Niu.Columns().Status).
		Value()
	if statusErr != nil {
		t.Fatalf("reload niu status failed: %v", statusErr)
	}
	status := statusVar.String()
	if status != cattlesvc.NiuStatusActive.String() {
		t.Fatalf("expected niu flipped to active, got %q", status)
	}
}

// TestActivateNearestNearbyInactiveNiuMatched verifies the server chooses the
// nearest visible inactive cattle when multiple candidates are within threshold.
func TestActivateNearestNearbyInactiveNiuMatched(t *testing.T) {
	ctx := context.Background()
	setupPostgreSQLActivationDB(t, ctx)
	svc := newActivationServiceForTest()
	past := time.Now().Add(-time.Hour)
	fartherID := insertNiuRow(t, ctx, do.Niu{
		Code: "NIU-ACT-FARTHER", NiuType: cattlesvc.NiuTypeCommon.String(),
		Lat: 30.0003, Lng: 103.0,
		OnlineAt: &past,
		Status:   cattlesvc.NiuStatusInactive.String(),
	})
	nearerID := insertNiuRow(t, ctx, do.Niu{
		Code: "NIU-ACT-NEARER", NiuType: cattlesvc.NiuTypeCommon.String(),
		Lat: 30.0, Lng: 103.0,
		OnlineAt: &past,
		Status:   cattlesvc.NiuStatusInactive.String(),
	})
	playerID := insertUserRow(t, ctx, do.User{Openid: "openid-nearest"})

	out, err := svc.Activate(ctx, playerID, &ActivateInput{Lat: 30.0, Lng: 103.0})
	if err != nil {
		t.Fatalf("nearest activation failed: %v", err)
	}
	if out.NiuId != nearerID {
		t.Fatalf("expected nearest niu %d, got %d (farther=%d)", nearerID, out.NiuId, fartherID)
	}
}

// TestActivateActiveNiuNotMatched verifies an already-active cattle is no longer
// matched by a GPS check-in that only targets unactivated cattle.
func TestActivateActiveNiuNotMatched(t *testing.T) {
	ctx := context.Background()
	setupPostgreSQLActivationDB(t, ctx)
	svc := newActivationServiceForTest()
	niuID := stageActivatableNiu(t, ctx)
	firstPlayer := insertUserRow(t, ctx, do.User{Openid: "openid-active-a"})
	secondPlayer := insertUserRow(t, ctx, do.User{Openid: "openid-active-b"})

	if _, err := svc.Activate(ctx, firstPlayer, &ActivateInput{Lat: 30.0, Lng: 103.0}); err != nil {
		t.Fatalf("first activation failed: %v", err)
	}
	_, err := svc.Activate(ctx, secondPlayer, &ActivateInput{Lat: 30.0, Lng: 103.0})
	assertBizCode(t, err, CodeNoNearbyNiu.RuntimeCode())

	count, countErr := dao.Activation.Ctx(ctx).Where(dao.Activation.Columns().NiuId, niuID).Count()
	if countErr != nil {
		t.Fatalf("count activations failed: %v", countErr)
	}
	if count != 1 {
		t.Fatalf("expected only the first activation row, got %d", count)
	}
}

// TestActivateConcurrentFirstActivatorUnique verifies two players activating the
// same unactivated cattle concurrently yield exactly one first-activator, with the
// row lock serializing the race.
func TestActivateConcurrentFirstActivatorUnique(t *testing.T) {
	ctx := context.Background()
	setupPostgreSQLActivationDB(t, ctx)
	svc := newActivationServiceForTest()
	niuID := stageActivatableNiu(t, ctx)
	playerA := insertUserRow(t, ctx, do.User{Openid: "openid-concurrent-a"})
	playerB := insertUserRow(t, ctx, do.User{Openid: "openid-concurrent-b"})

	var (
		wg      sync.WaitGroup
		mu      sync.Mutex
		results []*ActivateOutput
		errs    []error
	)
	activate := func(playerID int64) {
		defer wg.Done()
		out, err := svc.Activate(ctx, playerID, &ActivateInput{Lat: 30.0, Lng: 103.0})
		mu.Lock()
		defer mu.Unlock()
		results = append(results, out)
		errs = append(errs, err)
	}
	wg.Add(2)
	go activate(playerA)
	go activate(playerB)
	wg.Wait()

	firstCount := 0
	successCount := 0
	for i, err := range errs {
		if err != nil {
			assertBizCode(t, err, CodeNoNearbyNiu.RuntimeCode())
			continue
		}
		successCount++
		if results[i].IsFirst {
			firstCount++
		}
	}
	if successCount != 1 {
		t.Fatalf("expected exactly one successful activation, got %d", successCount)
	}
	if firstCount != 1 {
		t.Fatalf("expected exactly one first-activator, got %d", firstCount)
	}

	dbFirst, err := dao.Activation.Ctx(ctx).
		Where(dao.Activation.Columns().NiuId, niuID).
		Where(dao.Activation.Columns().IsFirst, firstActivatorFlag).
		Count()
	if err != nil {
		t.Fatalf("count first activations failed: %v", err)
	}
	if dbFirst != 1 {
		t.Fatalf("expected exactly one first-activator row, got %d", dbFirst)
	}
}

// TestActivateDailyLimitRejected verifies a second activation on the same day by
// the same player is rejected with CodeDailyLimitReached.
func TestActivateDailyLimitRejected(t *testing.T) {
	ctx := context.Background()
	setupPostgreSQLActivationDB(t, ctx)
	svc := newActivationServiceForTest()
	firstNiu := stageActivatableNiu(t, ctx)
	past := time.Now().Add(-time.Hour)
	secondNiu := insertNiuRow(t, ctx, do.Niu{
		Code: "NIU-ACT-002", NiuType: cattlesvc.NiuTypeCommon.String(),
		Lat: 30.0, Lng: 103.0,
		OnlineAt: &past,
		Status:   cattlesvc.NiuStatusInactive.String(),
	})
	playerID := insertUserRow(t, ctx, do.User{Openid: "openid-daily"})

	out, err := svc.Activate(ctx, playerID, &ActivateInput{Lat: 30.0, Lng: 103.0})
	if err != nil {
		t.Fatalf("first activation failed: %v", err)
	}
	if out.NiuId != firstNiu {
		t.Fatalf("expected first staged niu %d activated, got %d", firstNiu, out.NiuId)
	}
	_ = secondNiu
	_, err = svc.Activate(ctx, playerID, &ActivateInput{Lat: 30.0, Lng: 103.0})
	assertBizCode(t, err, CodeDailyLimitReached.RuntimeCode())
	attempts := activationAttemptsForPlayer(t, ctx, playerID)
	if len(attempts) != 1 {
		t.Fatalf("expected daily-limit rejection not to add attempt rows, got %d", len(attempts))
	}
}

// TestActivateNilInputRejected verifies an empty activation payload is rejected
// before any activation row is written.
func TestActivateNilInputRejected(t *testing.T) {
	ctx := context.Background()
	setupPostgreSQLActivationDB(t, ctx)
	svc := newActivationServiceForTest()
	playerID := insertUserRow(t, ctx, do.User{Openid: "openid-dup"})

	_, err := svc.Activate(ctx, playerID, nil)
	assertBizCode(t, err, CodeNoNearbyNiu.RuntimeCode())

	count, countErr := dao.Activation.Ctx(ctx).Count()
	if countErr != nil {
		t.Fatalf("count activations failed: %v", countErr)
	}
	if count != 0 {
		t.Fatalf("expected no activation rows, got %d", count)
	}
}
