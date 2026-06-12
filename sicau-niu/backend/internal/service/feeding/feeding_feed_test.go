// feeding_feed_test.go covers the DB-gated feeding action: a normal feed deducts
// the base amount and records coefficient 1.0, an iron cow within range raises the
// effect to coefficient 1.5, feeding a non-activated cattle is rejected, and an
// insufficient balance is rejected without recording a feeding.

package feeding

import (
	"context"
	"testing"

	"lina-plugin-sicau-niu/backend/internal/dao"
	"lina-plugin-sicau-niu/backend/internal/model/do"
	cattlesvc "lina-plugin-sicau-niu/backend/internal/service/cattle"
	"lina-plugin-sicau-niu/backend/internal/service/feeding/internal/ironlocation"
)

// stageActiveNiu inserts an activated cattle at the given anchor and returns its
// ID, so feeding's active-status check passes.
func stageActiveNiu(t *testing.T, ctx context.Context, code string, lat, lng float64) int64 {
	t.Helper()
	return insertNiuRow(t, ctx, do.Niu{
		Code:    code,
		NiuType: cattlesvc.NiuTypeCommon.String(),
		Lat:     lat, Lng: lng,
		Status: cattlesvc.NiuStatusActive.String(),
	})
}

// TestFeedNoBonus verifies feeding with no iron cow in range applies coefficient
// 1.0, deducts the base amount and keeps the balance equal to the ledger sum.
func TestFeedNoBonus(t *testing.T) {
	ctx := context.Background()
	setupPostgreSQLFeedingDB(t, ctx)
	svc := newFeedingServiceForTest(&fakeIronLocation{})
	insertQuoteRow(t, ctx, "校史金句", quoteEnabledOn)

	niuID := stageActiveNiu(t, ctx, "NIU-FEED-1", 30.0, 103.0)
	user := insertUserRow(t, ctx, "openid-feed-nobonus")
	creditGrass(t, ctx, user, 100)

	out, err := svc.Feed(ctx, user, &FeedInput{NiuId: niuID, BaseAmount: 10})
	if err != nil {
		t.Fatalf("feed failed: %v", err)
	}
	if out.CoefficientBasis != baseCoefficientBasis || out.EffectAmount != 10 || out.IsIronBonus {
		t.Fatalf("expected no bonus (basis 100, effect 10), got %+v", out)
	}
	if out.Balance != 90 {
		t.Fatalf("expected balance 90 after feeding 10, got %d", out.Balance)
	}
	if out.Quote == "" {
		t.Fatalf("expected a random quote, got empty")
	}
}

// TestFeedIronBonus verifies feeding with an iron cow within the threshold applies
// coefficient 1.5 and raises the effect.
func TestFeedIronBonus(t *testing.T) {
	ctx := context.Background()
	setupPostgreSQLFeedingDB(t, ctx)
	// The iron cow is placed at the exact cattle anchor, well within 12m.
	iron := &fakeIronLocation{positions: []*ironlocation.IronPosition{{IronID: 1, Lat: 30.0, Lng: 103.0}}}
	svc := newFeedingServiceForTest(iron)

	niuID := stageActiveNiu(t, ctx, "NIU-FEED-2", 30.0, 103.0)
	user := insertUserRow(t, ctx, "openid-feed-bonus")
	creditGrass(t, ctx, user, 100)

	out, err := svc.Feed(ctx, user, &FeedInput{NiuId: niuID, BaseAmount: 10})
	if err != nil {
		t.Fatalf("feed failed: %v", err)
	}
	if out.CoefficientBasis != ironBonusCoefficientBasis || out.EffectAmount != 15 || !out.IsIronBonus {
		t.Fatalf("expected iron bonus (basis 150, effect 15), got %+v", out)
	}
	if out.Balance != 90 {
		t.Fatalf("expected balance 90 (base deducted, not effect), got %d", out.Balance)
	}
}

// TestFeedIronOutOfRangeNoBonus verifies an iron cow beyond the threshold does not
// apply a bonus.
func TestFeedIronOutOfRangeNoBonus(t *testing.T) {
	ctx := context.Background()
	setupPostgreSQLFeedingDB(t, ctx)
	// The iron cow is ~1.1km away (0.01 degree latitude), far beyond 12m.
	iron := &fakeIronLocation{positions: []*ironlocation.IronPosition{{IronID: 1, Lat: 30.01, Lng: 103.0}}}
	svc := newFeedingServiceForTest(iron)

	niuID := stageActiveNiu(t, ctx, "NIU-FEED-3", 30.0, 103.0)
	user := insertUserRow(t, ctx, "openid-feed-far")
	creditGrass(t, ctx, user, 100)

	out, err := svc.Feed(ctx, user, &FeedInput{NiuId: niuID, BaseAmount: 10})
	if err != nil {
		t.Fatalf("feed failed: %v", err)
	}
	if out.IsIronBonus || out.CoefficientBasis != baseCoefficientBasis {
		t.Fatalf("expected no bonus for far iron cow, got %+v", out)
	}
}

// TestFeedDuplicateRequestRejected verifies a retry carrying an already-recorded
// request ID is rejected with CodeDuplicateRequest and deducts nothing further,
// while a fresh request ID still feeds normally.
func TestFeedDuplicateRequestRejected(t *testing.T) {
	ctx := context.Background()
	setupPostgreSQLFeedingDB(t, ctx)
	svc := newFeedingServiceForTest(&fakeIronLocation{})

	niuID := stageActiveNiu(t, ctx, "NIU-FEED-DUP", 30.0, 103.0)
	user := insertUserRow(t, ctx, "openid-feed-dup")
	creditGrass(t, ctx, user, 100)

	out, err := svc.Feed(ctx, user, &FeedInput{NiuId: niuID, BaseAmount: 10, RequestId: "req-feed-1"})
	if err != nil {
		t.Fatalf("first feed failed: %v", err)
	}
	if out.Balance != 90 {
		t.Fatalf("expected balance 90 after first feed, got %d", out.Balance)
	}

	_, err = svc.Feed(ctx, user, &FeedInput{NiuId: niuID, BaseAmount: 10, RequestId: "req-feed-1"})
	assertBizCode(t, err, CodeDuplicateRequest.RuntimeCode())

	feedCount, countErr := dao.Feeding.Ctx(ctx).Where(dao.Feeding.Columns().UserId, user).Count()
	if countErr != nil {
		t.Fatalf("count feedings failed: %v", countErr)
	}
	if feedCount != 1 {
		t.Fatalf("expected exactly one feeding row, got %d", feedCount)
	}

	out, err = svc.Feed(ctx, user, &FeedInput{NiuId: niuID, BaseAmount: 10, RequestId: "req-feed-2"})
	if err != nil {
		t.Fatalf("feed with fresh request ID failed: %v", err)
	}
	if out.Balance != 80 {
		t.Fatalf("expected balance 80 after second distinct feed, got %d", out.Balance)
	}
}

// TestFeedNotActiveRejected verifies feeding a not-yet-activated cattle is rejected
// and does not deduct grass.
func TestFeedNotActiveRejected(t *testing.T) {
	ctx := context.Background()
	setupPostgreSQLFeedingDB(t, ctx)
	svc := newFeedingServiceForTest(&fakeIronLocation{})

	niuID := insertNiuRow(t, ctx, do.Niu{
		Code:    "NIU-FEED-INACTIVE",
		NiuType: cattlesvc.NiuTypeCommon.String(),
		Lat:     30.0, Lng: 103.0,
		Status: cattlesvc.NiuStatusInactive.String(),
	})
	user := insertUserRow(t, ctx, "openid-feed-inactive")
	creditGrass(t, ctx, user, 100)

	_, err := svc.Feed(ctx, user, &FeedInput{NiuId: niuID, BaseAmount: 10})
	assertBizCode(t, err, CodeNiuNotActive.RuntimeCode())

	count, err := dao.Feeding.Ctx(ctx).Where(dao.Feeding.Columns().UserId, user).Count()
	if err != nil {
		t.Fatalf("count feeding failed: %v", err)
	}
	if count != 0 {
		t.Fatalf("expected no feeding recorded, got %d", count)
	}
}

// TestFeedInsufficientBalanceRejected verifies feeding beyond the balance is
// rejected and rolls back so no feeding record remains.
func TestFeedInsufficientBalanceRejected(t *testing.T) {
	ctx := context.Background()
	setupPostgreSQLFeedingDB(t, ctx)
	svc := newFeedingServiceForTest(&fakeIronLocation{})

	niuID := stageActiveNiu(t, ctx, "NIU-FEED-POOR", 30.0, 103.0)
	user := insertUserRow(t, ctx, "openid-feed-poor")
	creditGrass(t, ctx, user, 5)

	_, err := svc.Feed(ctx, user, &FeedInput{NiuId: niuID, BaseAmount: 10})
	if err == nil {
		t.Fatalf("expected insufficient-balance rejection, got nil")
	}

	count, err := dao.Feeding.Ctx(ctx).Where(dao.Feeding.Columns().UserId, user).Count()
	if err != nil {
		t.Fatalf("count feeding failed: %v", err)
	}
	if count != 0 {
		t.Fatalf("expected feeding rolled back, got %d records", count)
	}
}
