// grasssocial_steal_test.go covers the DB-gated steal action: a steal moves grass
// between both players in one transaction and notifies the target, a steal beyond
// the target's daily list is rejected, and the daily steal limit is enforced. A
// pure-logic check covers the deterministic daily-list derivation.

package grasssocial

import (
	"context"
	"testing"

	grasssvc "lina-plugin-sicau-niu/backend/internal/service/grass"
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

	out, err := svc.Steal(ctx, actor, &StealInput{TargetUserId: target})
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

	_, err = svc.Steal(ctx, actor, &StealInput{TargetUserId: outside})
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
		if _, err := svc.Steal(ctx, actor, &StealInput{TargetUserId: target}); err != nil {
			t.Fatalf("steal %d failed: %v", i, err)
		}
	}
	_, err := svc.Steal(ctx, actor, &StealInput{TargetUserId: target})
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
	grassService := grasssvc.New(nil, grasssvc.Config{CheckinMinAmount: 1, CheckinMaxAmount: 1})
	return New(grassService, nil, Config{
		StealDailyTargets: targets,
		StealDailyLimit:   2,
		StealMinAmount:    10,
		StealMaxAmount:    10,
		GiftDailyLimit:    2,
		GiftMinAmount:     12,
	})
}
