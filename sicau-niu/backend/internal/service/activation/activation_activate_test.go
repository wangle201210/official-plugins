// activation_activate_test.go covers the DB-gated LBS activation flow: GPS
// check-in matching, visible/inactive candidate gating, concurrent
// first-activator uniqueness, the per-day limit and on-activation card issuance.
// Each test is self-contained: it truncates the plugin tables, stages its own
// cattle/cards and asserts only on its own rows.

package activation

import (
	"context"
	"reflect"
	"strconv"
	"strings"
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
	_, err := svc.Activate(ctx, playerID, &ActivateInput{RequestID: "activation-no-nearby", Lat: 30.01, Lng: 103.0})
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

			_, err := svc.Activate(ctx, playerID, &ActivateInput{
				RequestID: "activation-invisible-" + strconv.Itoa(i),
				Lat:       30.0,
				Lng:       103.0,
			})
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

	out, err := svc.Activate(ctx, playerID, &ActivateInput{RequestID: "activation-first", Lat: 30.0, Lng: 103.0})
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

	out, err := svc.Activate(ctx, playerID, &ActivateInput{RequestID: "activation-nearest", Lat: 30.0, Lng: 103.0})
	if err != nil {
		t.Fatalf("nearest activation failed: %v", err)
	}
	if out.NiuId != nearerID {
		t.Fatalf("expected nearest niu %d, got %d (farther=%d)", nearerID, out.NiuId, fartherID)
	}
}

// TestActivateActiveNiuAllowsLaterVisitor verifies shared active state records a
// later player's own visit and issues the same cattle main card.
func TestActivateActiveNiuAllowsLaterVisitor(t *testing.T) {
	ctx := context.Background()
	setupPostgreSQLActivationDB(t, ctx)
	svc := newActivationServiceForTest()
	niuID := stageActivatableNiu(t, ctx)
	insertCardRow(t, ctx, do.Card{NiuId: niuID, Category: "event", Title: "后来者卡片", Content: "同行者也能获得"})
	firstPlayer := insertUserRow(t, ctx, do.User{Openid: "openid-active-a"})
	secondPlayer := insertUserRow(t, ctx, do.User{Openid: "openid-active-b"})

	if _, err := svc.Activate(ctx, firstPlayer, &ActivateInput{RequestID: "activation-active-first", Lat: 30.0, Lng: 103.0}); err != nil {
		t.Fatalf("first activation failed: %v", err)
	}
	out, err := svc.Activate(ctx, secondPlayer, &ActivateInput{RequestID: "activation-active-second", Lat: 30.0, Lng: 103.0})
	if err != nil {
		t.Fatalf("later activation failed: %v", err)
	}
	if out.IsFirst || out.OrderNo != 2 {
		t.Fatalf("expected later visitor order 2, got isFirst=%v order=%d", out.IsFirst, out.OrderNo)
	}
	if out.Card == nil || out.Card.Title != "后来者卡片" {
		t.Fatalf("expected later visitor main card, got %+v", out.Card)
	}

	count, countErr := dao.Activation.Ctx(ctx).Where(dao.Activation.Columns().NiuId, niuID).Count()
	if countErr != nil {
		t.Fatalf("count activations failed: %v", countErr)
	}
	if count != 2 {
		t.Fatalf("expected first and later activation rows, got %d", count)
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
	activate := func(playerID int64, requestID string) {
		defer wg.Done()
		out, err := svc.Activate(ctx, playerID, &ActivateInput{RequestID: requestID, Lat: 30.0, Lng: 103.0})
		mu.Lock()
		defer mu.Unlock()
		results = append(results, out)
		errs = append(errs, err)
	}
	wg.Add(2)
	go activate(playerA, "activation-concurrent-first-a")
	go activate(playerB, "activation-concurrent-first-b")
	wg.Wait()

	firstCount := 0
	successCount := 0
	for i, err := range errs {
		if err != nil {
			t.Fatalf("concurrent activation failed: %v", err)
		}
		successCount++
		if results[i].IsFirst {
			firstCount++
		}
	}
	if successCount != 2 {
		t.Fatalf("expected both players to activate successfully, got %d", successCount)
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

	out, err := svc.Activate(ctx, playerID, &ActivateInput{RequestID: "activation-daily-first", Lat: 30.0, Lng: 103.0})
	if err != nil {
		t.Fatalf("first activation failed: %v", err)
	}
	if out.NiuId != firstNiu {
		t.Fatalf("expected first staged niu %d activated, got %d", firstNiu, out.NiuId)
	}
	_ = secondNiu
	_, err = svc.Activate(ctx, playerID, &ActivateInput{RequestID: "activation-daily-second", Lat: 30.0, Lng: 103.0})
	assertBizCode(t, err, CodeDailyLimitReached.RuntimeCode())
	attempts := activationAttemptsForPlayer(t, ctx, playerID)
	if len(attempts) != 1 {
		t.Fatalf("expected daily-limit rejection not to add attempt rows, got %d", len(attempts))
	}
}

func TestActivateConcurrentDistinctRequestsReturnDailyLimit(t *testing.T) {
	ctx := context.Background()
	setupPostgreSQLActivationDB(t, ctx)
	svc := newActivationServiceForTest()
	stageActivatableNiu(t, ctx)
	playerID := insertUserRow(t, ctx, do.User{Openid: "openid-concurrent-daily"})

	var (
		wg   sync.WaitGroup
		mu   sync.Mutex
		errA error
		errB error
	)
	wg.Add(2)
	go func() {
		defer wg.Done()
		_, err := svc.Activate(ctx, playerID, &ActivateInput{RequestID: "activation-concurrent-a", Lat: 30.0, Lng: 103.0})
		mu.Lock()
		errA = err
		mu.Unlock()
	}()
	go func() {
		defer wg.Done()
		_, err := svc.Activate(ctx, playerID, &ActivateInput{RequestID: "activation-concurrent-b", Lat: 30.0, Lng: 103.0})
		mu.Lock()
		errB = err
		mu.Unlock()
	}()
	wg.Wait()

	if errA == nil && errB != nil {
		assertBizCode(t, errB, CodeDailyLimitReached.RuntimeCode())
	} else if errB == nil && errA != nil {
		assertBizCode(t, errA, CodeDailyLimitReached.RuntimeCode())
	} else {
		t.Fatalf("expected one success and one daily-limit error, got errA=%v errB=%v", errA, errB)
	}
}

func TestActivateRequestReplayReturnsExactResponse(t *testing.T) {
	ctx := context.Background()
	setupPostgreSQLActivationDB(t, ctx)
	svc := newActivationServiceForTest()
	stageActivatableNiu(t, ctx)
	playerID := insertUserRow(t, ctx, do.User{Openid: "openid-activation-replay"})
	in := &ActivateInput{RequestID: "activation-replay-1", Lat: 30.0, Lng: 103.0}

	first, err := svc.Activate(ctx, playerID, in)
	if err != nil {
		t.Fatalf("first activation failed: %v", err)
	}
	replayed, err := svc.Activate(ctx, playerID, in)
	if err != nil {
		t.Fatalf("activation replay failed: %v", err)
	}
	if !reflect.DeepEqual(first, replayed) {
		t.Fatalf("activation replay changed response: first=%+v replay=%+v", first, replayed)
	}
}

func TestRevokeRepairsCurrentFirstStateWithoutChangingReplay(t *testing.T) {
	ctx := context.Background()
	setupPostgreSQLActivationDB(t, ctx)
	svc := newActivationServiceForTest()
	niuID := stageActivatableNiu(t, ctx)
	firstPlayer := insertUserRow(t, ctx, do.User{Openid: "openid-revoke-first"})
	secondPlayer := insertUserRow(t, ctx, do.User{Openid: "openid-revoke-second"})

	first, err := svc.Activate(ctx, firstPlayer, &ActivateInput{RequestID: "activation-revoke-first", Lat: 30.0, Lng: 103.0})
	if err != nil {
		t.Fatalf("first activation failed: %v", err)
	}
	second, err := svc.Activate(ctx, secondPlayer, &ActivateInput{RequestID: "activation-revoke-second", Lat: 30.0, Lng: 103.0})
	if err != nil {
		t.Fatalf("second activation failed: %v", err)
	}
	if second.IsFirst {
		t.Fatal("second activation unexpectedly started as first")
	}
	if err = svc.Revoke(ctx, first.Seq); err != nil {
		t.Fatalf("revoke first activation failed: %v", err)
	}
	promoted, err := svc.Activate(ctx, secondPlayer, &ActivateInput{RequestID: "activation-revoke-second", Lat: 30.0, Lng: 103.0})
	if err != nil {
		t.Fatalf("replay promoted activation failed: %v", err)
	}
	if !reflect.DeepEqual(second, promoted) {
		t.Fatalf("operational repair rewrote the first successful replay: before=%+v after=%+v", second, promoted)
	}
	var repaired *entitymodel.Activation
	if err = dao.Activation.Ctx(ctx).Where(do.Activation{Id: second.Seq}).Scan(&repaired); err != nil || repaired == nil || repaired.IsFirst != firstActivatorFlag {
		t.Fatalf("current first-activator state was not repaired: row=%+v err=%v", repaired, err)
	}
	revokedReplay, err := svc.Activate(ctx, firstPlayer, &ActivateInput{RequestID: "activation-revoke-first", Lat: 30.0, Lng: 103.0})
	if err != nil || !reflect.DeepEqual(first, revokedReplay) {
		t.Fatalf("revoked request was re-executed instead of stably replayed: first=%+v replay=%+v err=%v", first, revokedReplay, err)
	}
	activeCount, err := dao.Activation.Ctx(ctx).Where(do.Activation{UserId: firstPlayer}).Count()
	if err != nil || activeCount != 0 {
		t.Fatalf("revoked activation became effective again: count=%d err=%v", activeCount, err)
	}
	if err = svc.Revoke(ctx, second.Seq); err != nil {
		t.Fatalf("revoke last activation failed: %v", err)
	}
	status, err := dao.Niu.Ctx(ctx).Where(do.Niu{Id: niuID}).Fields(dao.Niu.Columns().Status).Value()
	if err != nil || status.String() != cattlesvc.NiuStatusInactive.String() {
		t.Fatalf("last revoke did not restore inactive status: status=%q err=%v", status.String(), err)
	}
}

// insertAttemptRow inserts one activation-attempt audit row directly for test
// setup, staging a player's prior check-in history.
func insertAttemptRow(t *testing.T, ctx context.Context, playerID int64, lat, lng float64, attemptedAt time.Time) {
	t.Helper()
	_, err := dao.ActivationAttempt.Ctx(ctx).Data(do.ActivationAttempt{
		UserId:      playerID,
		Result:      string(activationAttemptNoNearby),
		Lat:         lat,
		Lng:         lng,
		AttemptedAt: &attemptedAt,
	}).Insert()
	if err != nil {
		t.Fatalf("insert attempt row failed: %v", err)
	}
}

// TestActivateAttemptLimitRejected verifies a check-in beyond the daily attempt
// quota (failures count) is rejected with CodeAttemptLimitReached and writes no
// further attempt rows.
func TestActivateAttemptLimitRejected(t *testing.T) {
	ctx := context.Background()
	setupPostgreSQLActivationDB(t, ctx)
	svc := newActivationServiceForTest()
	stageActivatableNiu(t, ctx)
	playerID := insertUserRow(t, ctx, do.User{Openid: "openid-attempt-limit"})

	// Fill today's quota with prior attempts at the same spot so the speed guard
	// stays quiet and only the quota triggers.
	for i := 0; i < defaultDailyAttemptLimit; i++ {
		insertAttemptRow(t, ctx, playerID, 30.0, 103.0, time.Now().Add(-time.Duration(i+1)*time.Minute))
	}

	_, err := svc.Activate(ctx, playerID, &ActivateInput{RequestID: "activation-attempt-limit", Lat: 30.0, Lng: 103.0})
	assertBizCode(t, err, CodeAttemptLimitReached.RuntimeCode())

	attempts := activationAttemptsForPlayer(t, ctx, playerID)
	if len(attempts) != defaultDailyAttemptLimit {
		t.Fatalf("expected quota rejection not to add attempt rows, got %d", len(attempts))
	}
	count, countErr := dao.Activation.Ctx(ctx).Count()
	if countErr != nil {
		t.Fatalf("count activations failed: %v", countErr)
	}
	if count != 0 {
		t.Fatalf("expected no activation rows, got %d", count)
	}
}

// TestActivateConcurrentAttemptsRespectQuota verifies the player row lock also
// protects the failed-attempt quota: with one slot left, two distinct concurrent
// requests produce one recorded attempt and one quota error.
func TestActivateConcurrentAttemptsRespectQuota(t *testing.T) {
	ctx := context.Background()
	setupPostgreSQLActivationDB(t, ctx)
	svc := newActivationServiceForTest()
	playerID := insertUserRow(t, ctx, do.User{Openid: "openid-attempt-concurrent"})

	for i := 0; i < defaultDailyAttemptLimit-1; i++ {
		insertAttemptRow(t, ctx, playerID, 30.0, 103.0, time.Now().Add(-time.Duration(i+1)*time.Minute))
	}

	errs := make(chan error, 2)
	var wg sync.WaitGroup
	for _, requestID := range []string{"attempt-concurrent-a", "attempt-concurrent-b"} {
		wg.Add(1)
		go func(requestID string) {
			defer wg.Done()
			_, err := svc.Activate(ctx, playerID, &ActivateInput{RequestID: requestID, Lat: 30.0, Lng: 103.0})
			errs <- err
		}(requestID)
	}
	wg.Wait()
	close(errs)

	noNearby := 0
	limited := 0
	for err := range errs {
		if isActivationBizCode(err, CodeNoNearbyNiu) {
			noNearby++
			continue
		}
		if isActivationBizCode(err, CodeAttemptLimitReached) {
			limited++
			continue
		}
		t.Fatalf("unexpected concurrent activation error: %v", err)
	}
	if noNearby != 1 || limited != 1 {
		t.Fatalf("expected one recorded failure and one quota error, noNearby=%d limited=%d", noNearby, limited)
	}
	attempts := activationAttemptsForPlayer(t, ctx, playerID)
	if len(attempts) != defaultDailyAttemptLimit {
		t.Fatalf("expected attempt count capped at %d, got %d", defaultDailyAttemptLimit, len(attempts))
	}
}

// TestActivateSpeedAnomalyRejected verifies a check-in implying implausible
// movement speed since the previous attempt is rejected with CodeSpeedAnomaly and
// recorded as a speed_anomaly risk row.
func TestActivateSpeedAnomalyRejected(t *testing.T) {
	ctx := context.Background()
	setupPostgreSQLActivationDB(t, ctx)
	svc := newActivationServiceForTest()
	niuID := stageActivatableNiu(t, ctx)
	playerID := insertUserRow(t, ctx, do.User{Openid: "openid-speed"})

	// Previous check-in ~5.5km away just 5 seconds ago: >1000 m/s, far over 25 m/s.
	insertAttemptRow(t, ctx, playerID, 30.05, 103.0, time.Now().Add(-5*time.Second))

	_, err := svc.Activate(ctx, playerID, &ActivateInput{RequestID: "activation-speed", Lat: 30.0, Lng: 103.0})
	assertBizCode(t, err, CodeSpeedAnomaly.RuntimeCode())

	attempts := activationAttemptsForPlayer(t, ctx, playerID)
	if len(attempts) != 2 {
		t.Fatalf("expected the staged attempt plus one speed_anomaly risk row, got %d", len(attempts))
	}
	if attempts[1].Result != string(activationAttemptSpeedAnomaly) {
		t.Fatalf("expected speed_anomaly risk row, got %q", attempts[1].Result)
	}
	count, countErr := dao.Activation.Ctx(ctx).Where(dao.Activation.Columns().NiuId, niuID).Count()
	if countErr != nil {
		t.Fatalf("count activations failed: %v", countErr)
	}
	if count != 0 {
		t.Fatalf("expected no activation rows, got %d", count)
	}
}

// TestActivateSlowMovementPasses verifies a plausible movement between attempts
// passes the speed guard and the check-in proceeds to matching.
func TestActivateSlowMovementPasses(t *testing.T) {
	ctx := context.Background()
	setupPostgreSQLActivationDB(t, ctx)
	svc := newActivationServiceForTest()
	niuID := stageActivatableNiu(t, ctx)
	playerID := insertUserRow(t, ctx, do.User{Openid: "openid-slow"})

	// Previous check-in ~110m away one hour ago: ~0.03 m/s, far under 25 m/s.
	insertAttemptRow(t, ctx, playerID, 30.001, 103.0, time.Now().Add(-time.Hour))

	out, err := svc.Activate(ctx, playerID, &ActivateInput{RequestID: "activation-slow", Lat: 30.0, Lng: 103.0})
	if err != nil {
		t.Fatalf("expected slow movement to pass the speed guard, got %v", err)
	}
	if out.NiuId != niuID {
		t.Fatalf("expected matched niu %d, got %d", niuID, out.NiuId)
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
	assertBizCode(t, err, CodeRequestIDRequired.RuntimeCode())

	count, countErr := dao.Activation.Ctx(ctx).Count()
	if countErr != nil {
		t.Fatalf("count activations failed: %v", countErr)
	}
	if count != 0 {
		t.Fatalf("expected no activation rows, got %d", count)
	}
}

// TestActivateRequiresRequestID verifies internal callers cannot bypass the
// mandatory idempotency-key contract enforced by the HTTP DTO.
func TestActivateRequiresRequestID(t *testing.T) {
	for _, requestID := range []string{"", "   ", strings.Repeat("r", 65)} {
		_, err := activationRequestID(1, &ActivateInput{RequestID: requestID})
		assertBizCode(t, err, CodeRequestIDRequired.RuntimeCode())
	}
	if got, err := activationRequestID(1, &ActivateInput{RequestID: strings.Repeat("牛", 64)}); err != nil || got == "" {
		t.Fatalf("64 Unicode request-id characters should pass validation: got=%q err=%v", got, err)
	}
}
