// activation_activate_test.go covers the DB-gated LBS activation flow: the
// distance gate, shared-pool first-activator, concurrent first-activator
// uniqueness, the per-day limit, the no-duplicate rule and on-activation card
// issuance. Each test is self-contained: it truncates the plugin tables, stages
// its own cattle/cards and asserts only on its own rows.

package activation

import (
	"context"
	"sync"
	"testing"

	"lina-plugin-sicau-niu/backend/internal/dao"
	"lina-plugin-sicau-niu/backend/internal/model/do"
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
	return insertNiuRow(t, ctx, do.Niu{
		Code:         "NIU-ACT-001",
		NiuType:      cattlesvc.NiuTypeCommon.String(),
		Lat:          30.0,
		Lng:          103.0,
		ReleaseStage: cattlesvc.ReleaseStageMain.String(),
		Status:       cattlesvc.NiuStatusInactive.String(),
	})
}

// TestActivateOutOfRangeRejected verifies a location beyond the LBS threshold is
// rejected with CodeOutOfRange and no activation row is written.
func TestActivateOutOfRangeRejected(t *testing.T) {
	ctx := context.Background()
	setupPostgreSQLActivationDB(t, ctx)
	svc := newActivationServiceForTest()
	niuID := stageActivatableNiu(t, ctx)
	playerID := insertUserRow(t, ctx, do.User{Openid: "openid-far"})

	// ~0.01 degree latitude offset is well over a kilometer, far outside 50m.
	_, err := svc.Activate(ctx, playerID, &ActivateInput{NiuId: niuID, Lat: 30.01, Lng: 103.0})
	assertBizCode(t, err, CodeOutOfRange.RuntimeCode())

	count, countErr := dao.Activation.Ctx(ctx).Where(dao.Activation.Columns().NiuId, niuID).Count()
	if countErr != nil {
		t.Fatalf("count activations failed: %v", countErr)
	}
	if count != 0 {
		t.Fatalf("expected no activation rows, got %d", count)
	}
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

	out, err := svc.Activate(ctx, playerID, &ActivateInput{NiuId: niuID, Lat: 30.0, Lng: 103.0})
	if err != nil {
		t.Fatalf("first activation failed: %v", err)
	}
	if !out.IsFirst || out.OrderNo != 1 {
		t.Fatalf("expected first activator with order 1, got isFirst=%v order=%d", out.IsFirst, out.OrderNo)
	}
	if out.Card == nil || out.Card.Title != "川农大精神" {
		t.Fatalf("expected issued main card, got %+v", out.Card)
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

// TestActivateLaterActivatorIncrementsOrder verifies a second player activating an
// already-active cattle is recorded as a non-first activator with order 2.
func TestActivateLaterActivatorIncrementsOrder(t *testing.T) {
	ctx := context.Background()
	setupPostgreSQLActivationDB(t, ctx)
	svc := newActivationServiceForTest()
	niuID := stageActivatableNiu(t, ctx)
	firstPlayer := insertUserRow(t, ctx, do.User{Openid: "openid-a"})
	secondPlayer := insertUserRow(t, ctx, do.User{Openid: "openid-b"})

	if _, err := svc.Activate(ctx, firstPlayer, &ActivateInput{NiuId: niuID, Lat: 30.0, Lng: 103.0}); err != nil {
		t.Fatalf("first activation failed: %v", err)
	}
	out, err := svc.Activate(ctx, secondPlayer, &ActivateInput{NiuId: niuID, Lat: 30.0, Lng: 103.0})
	if err != nil {
		t.Fatalf("second activation failed: %v", err)
	}
	if out.IsFirst || out.OrderNo != 2 {
		t.Fatalf("expected later activator with order 2, got isFirst=%v order=%d", out.IsFirst, out.OrderNo)
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
		out, err := svc.Activate(ctx, playerID, &ActivateInput{NiuId: niuID, Lat: 30.0, Lng: 103.0})
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
	for i, err := range errs {
		if err != nil {
			t.Fatalf("concurrent activation %d failed: %v", i, err)
		}
		if results[i].IsFirst {
			firstCount++
		}
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
	secondNiu := insertNiuRow(t, ctx, do.Niu{
		Code: "NIU-ACT-002", NiuType: cattlesvc.NiuTypeCommon.String(),
		Lat: 30.0, Lng: 103.0,
		ReleaseStage: cattlesvc.ReleaseStageMain.String(),
		Status:       cattlesvc.NiuStatusInactive.String(),
	})
	playerID := insertUserRow(t, ctx, do.User{Openid: "openid-daily"})

	if _, err := svc.Activate(ctx, playerID, &ActivateInput{NiuId: firstNiu, Lat: 30.0, Lng: 103.0}); err != nil {
		t.Fatalf("first activation failed: %v", err)
	}
	_, err := svc.Activate(ctx, playerID, &ActivateInput{NiuId: secondNiu, Lat: 30.0, Lng: 103.0})
	assertBizCode(t, err, CodeDailyLimitReached.RuntimeCode())
}

// TestActivateDuplicateNiuRejected verifies a repeat activation of the same cattle
// by the same player is rejected with CodeAlreadyActivated. The repeat attempt is
// staged on a different day by pre-seeding the prior activation directly so the
// daily limit does not mask the duplicate rule.
func TestActivateDuplicateNiuRejected(t *testing.T) {
	ctx := context.Background()
	setupPostgreSQLActivationDB(t, ctx)
	svc := newActivationServiceForTest()
	niuID := stageActivatableNiu(t, ctx)
	playerID := insertUserRow(t, ctx, do.User{Openid: "openid-dup"})

	// Seed a prior activation of this cattle on an earlier day so today's attempt
	// trips the no-duplicate rule rather than the daily limit.
	if _, err := dao.Activation.Ctx(ctx).Data(do.Activation{
		UserId:       playerID,
		NiuId:        niuID,
		ActivityDate: "2000-01-01",
		IsFirst:      firstActivatorFlag,
		OrderNo:      1,
	}).Insert(); err != nil {
		t.Fatalf("seed prior activation failed: %v", err)
	}

	_, err := svc.Activate(ctx, playerID, &ActivateInput{NiuId: niuID, Lat: 30.0, Lng: 103.0})
	assertBizCode(t, err, CodeAlreadyActivated.RuntimeCode())
}

// TestActivateMissingCattleRejected verifies activating a non-existent cattle is
// rejected with CodeNiuNotFound.
func TestActivateMissingCattleRejected(t *testing.T) {
	ctx := context.Background()
	setupPostgreSQLActivationDB(t, ctx)
	svc := newActivationServiceForTest()
	playerID := insertUserRow(t, ctx, do.User{Openid: "openid-missing"})

	_, err := svc.Activate(ctx, playerID, &ActivateInput{NiuId: 999999, Lat: 30.0, Lng: 103.0})
	assertBizCode(t, err, CodeNiuNotFound.RuntimeCode())
}
