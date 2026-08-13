// grasssocial_steal_test.go covers the DB-gated steal action: a steal moves grass
// between both players in one transaction and notifies the target, a steal beyond
// the target's daily list is rejected, and the daily steal limit is enforced. A
// pure-logic check covers the deterministic daily-list derivation.

package grasssocial

import (
	"context"
	"fmt"
	"reflect"
	"sync"
	"testing"
	"time"

	"lina-plugin-sicau-niu/backend/internal/activityday"
	"lina-plugin-sicau-niu/backend/internal/dao"
	"lina-plugin-sicau-niu/backend/internal/model/do"
	grasssvc "lina-plugin-sicau-niu/backend/internal/service/grass"
	rulessvc "lina-plugin-sicau-niu/backend/internal/service/rules"
)

// TestStealMovesGrassAndNotifies verifies a steal debits the target, credits the
// actor in one transaction and writes a stolen message to the target.
func TestStealMovesGrassAndNotifies(t *testing.T) {
	ctx := context.Background()
	setupPostgreSQLSocialDB(t, ctx)
	svc := newSocialServiceForTest()

	actor := insertUserRow(t, ctx, "openid-steal-actor")
	target := insertUserRow(t, ctx, "openid-steal-target")
	seedGrass(t, ctx, target, 100)

	out, err := svc.Steal(ctx, actor, &StealInput{TargetUserId: target, RequestId: "req-steal-normal"})
	if err != nil {
		t.Fatalf("steal failed: %v", err)
	}
	if out.Amount != 10 {
		t.Fatalf("expected fixed steal amount 10, got %d", out.Amount)
	}
	if got := balanceOf(t, ctx, actor); got != 10 {
		t.Fatalf("expected actor balance 10, got %d", got)
	}
	if got := balanceOf(t, ctx, target); got != 90 {
		t.Fatalf("expected target balance 90, got %d", got)
	}
	if n := inboxCount(t, ctx, target, MsgTypeStolen); n != 1 {
		t.Fatalf("expected 1 stolen notification, got %d", n)
	}
}

// TestStealUsesOneLockedRuleSnapshot verifies target authorization, daily limit
// and amount selection all consume one rule version after player locking. A
// replay reads no rules and the next new write observes the next version.
func TestStealUsesOneLockedRuleSnapshot(t *testing.T) {
	ctx := context.Background()
	setupPostgreSQLSocialDB(t, ctx)
	rules := newSequenceRulesService(
		rulessvc.SocialRules{
			StealDailyTargets: 2,
			StealDailyLimit:   2,
			StealMinAmount:    10,
			StealMaxAmount:    10,
		},
		rulessvc.SocialRules{
			StealDailyTargets: 1,
			StealDailyLimit:   1,
			StealMinAmount:    99,
			StealMaxAmount:    99,
		},
	)
	svc := newSocialServiceWithRules(rules)

	actor := insertUserRow(t, ctx, "openid-steal-snapshot-actor")
	candidates := []int64{
		insertUserRow(t, ctx, "openid-steal-snapshot-target-a"),
		insertUserRow(t, ctx, "openid-steal-snapshot-target-b"),
	}
	order := deterministicOrder(len(candidates), dailySeed(actor, activityday.Today()))
	target := candidates[order[1]]
	seedGrass(t, ctx, target, 100)
	if _, err := dao.Steal.Ctx(ctx).Data(do.Steal{
		ActorUserId:  actor,
		TargetUserId: target,
		Amount:       1,
		StealDate:    activityday.Today(),
		RequestId:    "req-steal-snapshot-prior",
	}).Insert(); err != nil {
		t.Fatalf("insert prior steal failed: %v", err)
	}

	first, err := svc.Steal(ctx, actor, &StealInput{TargetUserId: target, RequestId: "req-steal-snapshot"})
	if err != nil {
		t.Fatalf("steal with first rule snapshot failed: %v", err)
	}
	if first.Amount != 10 {
		t.Fatalf("expected first snapshot amount 10, got %d", first.Amount)
	}
	replayed, err := svc.Steal(ctx, actor, &StealInput{TargetUserId: target, RequestId: "req-steal-snapshot"})
	if err != nil || !reflect.DeepEqual(first, replayed) {
		t.Fatalf("steal replay changed the first result: first=%+v replay=%+v err=%v", first, replayed, err)
	}
	if got := rules.callCount(); got != 1 {
		t.Fatalf("expected one rule read for steal plus replay, got %d", got)
	}

	_, err = svc.Steal(ctx, actor, &StealInput{TargetUserId: target, RequestId: "req-steal-snapshot-next"})
	assertBizCode(t, err, CodeTargetNotStealable.RuntimeCode())
	if got := rules.callCount(); got != 2 {
		t.Fatalf("expected the next steal to read the next rule snapshot once, got %d total reads", got)
	}
}

// TestStealDuplicateRequestReplaysExactResult verifies a retry carrying an
// already-recorded request ID returns the first result and moves no further grass.
func TestStealDuplicateRequestReplaysExactResult(t *testing.T) {
	ctx := context.Background()
	setupPostgreSQLSocialDB(t, ctx)
	svc := newSocialServiceForTest()

	actor := insertUserRow(t, ctx, "openid-steal-dup-actor")
	target := insertUserRow(t, ctx, "openid-steal-dup-target")
	seedGrass(t, ctx, target, 100)

	out, err := svc.Steal(ctx, actor, &StealInput{TargetUserId: target, RequestId: "req-steal-1"})
	if err != nil {
		t.Fatalf("first steal failed: %v", err)
	}
	if out.Amount <= 0 {
		t.Fatalf("expected positive stolen amount, got %d", out.Amount)
	}

	replayed, err := svc.Steal(ctx, actor, &StealInput{TargetUserId: target, RequestId: "req-steal-1"})
	if err != nil {
		t.Fatalf("steal replay failed: %v", err)
	}
	if !reflect.DeepEqual(out, replayed) {
		t.Fatalf("expected exact first response on replay, first=%+v replay=%+v", out, replayed)
	}

	if got := balanceOf(t, ctx, actor); got != int64(out.Amount) {
		t.Fatalf("expected actor balance unchanged at %d after duplicate, got %d", out.Amount, got)
	}
	if got := balanceOf(t, ctx, target); got != 100-int64(out.Amount) {
		t.Fatalf("expected target balance unchanged at %d after duplicate, got %d", 100-int64(out.Amount), got)
	}
}

// TestStealConcurrentDailyLimitEnforced verifies distinct concurrent request IDs
// from one actor cannot both pass a one-action daily cap.
func TestStealConcurrentDailyLimitEnforced(t *testing.T) {
	ctx := context.Background()
	setupPostgreSQLSocialDB(t, ctx)
	svc := newSocialServiceWithLimits(1, 2)

	actor := insertUserRow(t, ctx, "openid-steal-concurrent-actor")
	target := insertUserRow(t, ctx, "openid-steal-concurrent-target")
	seedGrass(t, ctx, target, 100)

	errs := make(chan error, 2)
	var wg sync.WaitGroup
	for _, requestID := range []string{"req-steal-a", "req-steal-b"} {
		wg.Add(1)
		go func(requestID string) {
			defer wg.Done()
			_, err := svc.Steal(ctx, actor, &StealInput{TargetUserId: target, RequestId: requestID})
			errs <- err
		}(requestID)
	}
	wg.Wait()
	close(errs)

	successes := 0
	limitErrors := 0
	for err := range errs {
		if err == nil {
			successes++
			continue
		}
		if isBizCode(err, CodeStealLimitReached) {
			limitErrors++
			continue
		}
		t.Fatalf("unexpected concurrent steal error: %v", err)
	}
	if successes != 1 || limitErrors != 1 {
		t.Fatalf("expected one success and one limit error, successes=%d limits=%d", successes, limitErrors)
	}
	count, err := dao.Steal.Ctx(ctx).Where(dao.Steal.Columns().ActorUserId, actor).Count()
	if err != nil || count != 1 {
		t.Fatalf("expected one steal row, count=%d err=%v", count, err)
	}
}

// TestStealConcurrentSameRequestReplaysOnce verifies concurrent delivery of one
// idempotency key returns one stable outcome without a duplicate-key error.
func TestStealConcurrentSameRequestReplaysOnce(t *testing.T) {
	ctx := context.Background()
	setupPostgreSQLSocialDB(t, ctx)
	svc := newSocialServiceForTest()

	actor := insertUserRow(t, ctx, "openid-steal-replay-actor")
	target := insertUserRow(t, ctx, "openid-steal-replay-target")
	seedGrass(t, ctx, target, 100)

	results := make(chan *StealResult, 2)
	errs := make(chan error, 2)
	var wg sync.WaitGroup
	for i := 0; i < 2; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			out, err := svc.Steal(ctx, actor, &StealInput{TargetUserId: target, RequestId: "req-steal-same"})
			results <- out
			errs <- err
		}()
	}
	wg.Wait()
	close(results)
	close(errs)
	for err := range errs {
		if err != nil {
			t.Fatalf("concurrent steal replay failed: %v", err)
		}
	}
	var first *StealResult
	for out := range results {
		if first == nil {
			first = out
			continue
		}
		if !reflect.DeepEqual(first, out) {
			t.Fatalf("concurrent steal responses differ: first=%+v second=%+v", first, out)
		}
	}
	count, err := dao.Steal.Ctx(ctx).Where(dao.Steal.Columns().ActorUserId, actor).Count()
	if err != nil || count != 1 {
		t.Fatalf("expected one steal row, count=%d err=%v", count, err)
	}
}

// TestStealReciprocalTransfersUseStableLockOrder verifies two players stealing
// from each other concurrently complete without opposite-order account locks.
func TestStealReciprocalTransfersUseStableLockOrder(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	setupPostgreSQLSocialDB(t, ctx)
	svc := newSocialServiceForTest()
	first := insertUserRow(t, ctx, "openid-steal-reciprocal-a")
	second := insertUserRow(t, ctx, "openid-steal-reciprocal-b")
	seedGrass(t, ctx, first, 100)
	seedGrass(t, ctx, second, 100)

	errCh := make(chan error, 2)
	var wg sync.WaitGroup
	for _, call := range []struct {
		actor, target int64
		request       string
	}{{first, second, "req-steal-a-from-b"}, {second, first, "req-steal-b-from-a"}} {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, err := svc.Steal(ctx, call.actor, &StealInput{TargetUserId: call.target, RequestId: call.request})
			errCh <- err
		}()
	}
	wg.Wait()
	close(errCh)
	for err := range errCh {
		if err != nil {
			t.Fatalf("reciprocal steal failed: %v", err)
		}
	}
	if got := balanceOf(t, ctx, first); got != 100 {
		t.Fatalf("expected first balance 100 after reciprocal steals, got %d", got)
	}
	if got := balanceOf(t, ctx, second); got != 100 {
		t.Fatalf("expected second balance 100 after reciprocal steals, got %d", got)
	}
}

// TestStealTargetNotInListRejected verifies stealing from a player not in the
// actor's daily list is rejected. With more candidates than the list size, at
// least one other player is excluded.
func TestStealTargetNotInListRejected(t *testing.T) {
	ctx := context.Background()
	setupPostgreSQLSocialDB(t, ctx)
	// A small list size so most candidates are excluded from the daily list.
	svc := newSocialServiceWithTargets(1)

	actor := insertUserRow(t, ctx, "openid-steal-actor-2")
	// Create several candidates; the list of size 1 includes only one of them.
	candidates := make([]int64, 0, 5)
	for i := 0; i < 5; i++ {
		candidates = append(candidates, insertUserRow(t, ctx, "openid-cand-"+string(rune('a'+i))))
	}
	for _, c := range candidates {
		seedGrass(t, ctx, c, 100)
	}

	targets, err := svc.StealTargets(ctx, actor)
	if err != nil {
		t.Fatalf("steal targets failed: %v", err)
	}
	if len(targets) != 1 {
		t.Fatalf("expected exactly 1 daily target, got %d", len(targets))
	}
	inList := targets[0].UserId

	// Pick a candidate that is not the single listed target.
	var outside int64
	for _, c := range candidates {
		if c != inList {
			outside = c
			break
		}
	}

	_, err = svc.Steal(ctx, actor, &StealInput{TargetUserId: outside, RequestId: "req-steal-outside"})
	assertBizCode(t, err, CodeTargetNotStealable.RuntimeCode())
}

// TestStealDailyLimitEnforced verifies the actor cannot steal more than the daily
// limit. The fixture limit is 2.
func TestStealDailyLimitEnforced(t *testing.T) {
	ctx := context.Background()
	setupPostgreSQLSocialDB(t, ctx)
	svc := newSocialServiceForTest()

	actor := insertUserRow(t, ctx, "openid-steal-limit-actor")
	target := insertUserRow(t, ctx, "openid-steal-limit-target")
	seedGrass(t, ctx, target, 100)

	for i := 0; i < 2; i++ {
		if _, err := svc.Steal(ctx, actor, &StealInput{TargetUserId: target, RequestId: fmt.Sprintf("req-steal-limit-%d", i)}); err != nil {
			t.Fatalf("steal %d failed: %v", i, err)
		}
	}
	_, err := svc.Steal(ctx, actor, &StealInput{TargetUserId: target, RequestId: "req-steal-limit-over"})
	assertBizCode(t, err, CodeStealLimitReached.RuntimeCode())
}

// TestStealSelfRejected verifies stealing from oneself is rejected without a list
// lookup.
func TestStealSelfRejected(t *testing.T) {
	ctx := context.Background()
	svc := newSocialServiceForTest()
	_, err := svc.Steal(ctx, 7, &StealInput{TargetUserId: 7})
	assertBizCode(t, err, CodeSelfActionForbidden.RuntimeCode())
}

// TestStealRequiresRequestID verifies internal callers cannot bypass the public
// idempotency-key requirement.
func TestStealRequiresRequestID(t *testing.T) {
	svc := newSocialServiceForTest()
	_, err := svc.Steal(context.Background(), 1, &StealInput{TargetUserId: 2})
	assertBizCode(t, err, CodeRequestIDRequired.RuntimeCode())
}

// TestDeterministicDailyOrderStable verifies the deterministic order for a fixed
// seed is reproducible, the foundation of the recomputable daily list. This is a
// pure-logic check.
func TestDeterministicDailyOrderStable(t *testing.T) {
	seed := dailySeed(42, "2026-06-03")
	first := deterministicOrder(20, seed)
	second := deterministicOrder(20, seed)
	if len(first) != 20 {
		t.Fatalf("expected a permutation of length 20, got %d", len(first))
	}
	for i := range first {
		if first[i] != second[i] {
			t.Fatalf("deterministic order not stable at index %d: %d vs %d", i, first[i], second[i])
		}
	}
	// A different day yields a different seed and (almost surely) a different order.
	if dailySeed(42, "2026-06-04") == seed {
		t.Fatalf("expected a different seed for a different day")
	}
}

// newSocialServiceWithTargets builds a grass-social service with a custom daily
// stealable-list size and the fixture's deterministic steal amount and limits, so
// a test can force most candidates outside the daily list.
func newSocialServiceWithTargets(targets int) Service {
	return newSocialServiceWithConfig(targets, 2, 2)
}

func newSocialServiceWithLimits(stealLimit, giftLimit int) Service {
	return newSocialServiceWithConfig(12, stealLimit, giftLimit)
}

func newSocialServiceWithConfig(targets, stealLimit, giftLimit int) Service {
	grassService := grasssvc.New(nil, grasssvc.Config{CheckinMinAmount: 1, CheckinMaxAmount: 1})
	return New(grassService, nil, Config{
		StealDailyTargets: targets,
		StealDailyLimit:   stealLimit,
		StealMinAmount:    10,
		StealMaxAmount:    10,
		GiftDailyLimit:    giftLimit,
		GiftMinAmount:     12,
	})
}
