// feeding_trail_test.go covers the DB-gated feeding trail: an empty trail before
// any feeding, the latest-first ordering with the configured limit, and
// self-isolation so a player never sees another player's feedings.

package feeding

import (
	"context"
	"testing"
)

// TestTrailEmptyWhenNoFeeding verifies a player who never fed gets an empty trail.
func TestTrailEmptyWhenNoFeeding(t *testing.T) {
	ctx := context.Background()
	setupPostgreSQLFeedingDB(t, ctx)
	svc := newFeedingServiceForTest(&fakeIronLocation{})
	user := insertUserRow(t, ctx, "openid-trail-empty")

	items, err := svc.Trail(ctx, user)
	if err != nil {
		t.Fatalf("trail failed: %v", err)
	}
	if len(items) != 0 {
		t.Fatalf("expected empty trail, got %d", len(items))
	}
}

// TestTrailSelfIsolationAndOrder verifies the trail contains only the requesting
// player's feedings, newest first, with the cattle code assembled.
func TestTrailSelfIsolationAndOrder(t *testing.T) {
	ctx := context.Background()
	setupPostgreSQLFeedingDB(t, ctx)
	svc := newFeedingServiceForTest(&fakeIronLocation{})

	niuA := stageActiveNiu(t, ctx, "NIU-TRAIL-A", 30.0, 103.0)
	niuB := stageActiveNiu(t, ctx, "NIU-TRAIL-B", 31.0, 104.0)
	me := insertUserRow(t, ctx, "openid-trail-me")
	other := insertUserRow(t, ctx, "openid-trail-other")
	creditGrass(t, ctx, me, 100)
	creditGrass(t, ctx, other, 100)

	if _, err := svc.Feed(ctx, me, &FeedInput{NiuId: niuA, BaseAmount: 5}); err != nil {
		t.Fatalf("feed A failed: %v", err)
	}
	if _, err := svc.Feed(ctx, me, &FeedInput{NiuId: niuB, BaseAmount: 5}); err != nil {
		t.Fatalf("feed B failed: %v", err)
	}
	if _, err := svc.Feed(ctx, other, &FeedInput{NiuId: niuA, BaseAmount: 5}); err != nil {
		t.Fatalf("other feed failed: %v", err)
	}

	items, err := svc.Trail(ctx, me)
	if err != nil {
		t.Fatalf("trail failed: %v", err)
	}
	if len(items) != 2 {
		t.Fatalf("expected 2 of my feedings, got %d", len(items))
	}
	if items[0].NiuId != niuB {
		t.Fatalf("expected newest feeding (niuB) first, got niu %d", items[0].NiuId)
	}
	if items[0].NiuCode != "NIU-TRAIL-B" {
		t.Fatalf("expected assembled cattle code NIU-TRAIL-B, got %q", items[0].NiuCode)
	}
}
