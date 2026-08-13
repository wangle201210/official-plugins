// feeding_feed_test.go covers the DB-gated feeding action: a normal feed deducts
// the base amount and records coefficient 1.0, an iron cow within range raises the
// effect to coefficient 1.5, feeding a non-activated cattle is rejected, and an
// insufficient balance is rejected without recording a feeding.

package feeding

import (
	"context"
	"errors"
	"reflect"
	"sync"
	"testing"

	"github.com/gogf/gf/v2/database/gdb"

	"lina-plugin-sicau-niu/backend/internal/dao"
	"lina-plugin-sicau-niu/backend/internal/model/do"
	cattlesvc "lina-plugin-sicau-niu/backend/internal/service/cattle"
	"lina-plugin-sicau-niu/backend/internal/service/feeding/internal/ironlocation"
	grasssvc "lina-plugin-sicau-niu/backend/internal/service/grass"
)

type failAfterFirstIronLocation struct {
	mu         sync.Mutex
	positions  []*ironlocation.IronPosition
	err        error
	secondCall chan struct{}
	calls      int
}

func (g *failAfterFirstIronLocation) Positions(context.Context) ([]*ironlocation.IronPosition, error) {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.calls++
	if g.calls == 1 {
		return g.positions, nil
	}
	if g.calls == 2 {
		close(g.secondCall)
	}
	return nil, g.err
}

type blockingGrassService struct {
	grasssvc.Service
	entered chan struct{}
	release chan struct{}
	once    sync.Once
}

func (s *blockingGrassService) ApplyDelta(ctx context.Context, tx gdb.TX, userID int64, delta int64, txnType grasssvc.TxnType, refID int64) (int64, error) {
	s.once.Do(func() { close(s.entered) })
	<-s.release
	return s.Service.ApplyDelta(ctx, tx, userID, delta, txnType, refID)
}

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

	out, err := svc.Feed(ctx, user, &FeedInput{NiuId: niuID, BaseAmount: 10, RequestId: "req-feed-nobonus"})
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

	out, err := svc.Feed(ctx, user, &FeedInput{NiuId: niuID, BaseAmount: 10, RequestId: "req-feed-bonus"})
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

	out, err := svc.Feed(ctx, user, &FeedInput{NiuId: niuID, BaseAmount: 10, RequestId: "req-feed-far"})
	if err != nil {
		t.Fatalf("feed failed: %v", err)
	}
	if out.IsIronBonus || out.CoefficientBasis != baseCoefficientBasis {
		t.Fatalf("expected no bonus for far iron cow, got %+v", out)
	}
}

// TestFeedDuplicateRequestReplaysExactResult verifies a retry carrying an
// already-recorded request ID returns the first successful response exactly and
// deducts nothing further, even when mutable catalog data changes afterwards.
func TestFeedDuplicateRequestReplaysExactResult(t *testing.T) {
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

	if _, err = dao.Niu.Ctx(ctx).Where(dao.Niu.Columns().Id, niuID).Data(do.Niu{Name: "changed name"}).Update(); err != nil {
		t.Fatalf("update cattle name failed: %v", err)
	}
	replayed, err := svc.Feed(ctx, user, &FeedInput{NiuId: niuID, BaseAmount: 10, RequestId: "req-feed-1"})
	if err != nil {
		t.Fatalf("replay feed failed: %v", err)
	}
	if !reflect.DeepEqual(out, replayed) {
		t.Fatalf("expected exact first response on replay, first=%+v replay=%+v", out, replayed)
	}

	feedCount, countErr := dao.Feeding.Ctx(ctx).Where(dao.Feeding.Columns().UserId, user).Count()
	if countErr != nil {
		t.Fatalf("count feedings failed: %v", countErr)
	}
	if feedCount != 1 {
		t.Fatalf("expected exactly one feeding row, got %d", feedCount)
	}
	if balance := feedingBalanceOf(t, ctx, user); balance != 90 {
		t.Fatalf("expected replay to preserve balance 90, got %d", balance)
	}

	out, err = svc.Feed(ctx, user, &FeedInput{NiuId: niuID, BaseAmount: 10, RequestId: "req-feed-2"})
	if err != nil {
		t.Fatalf("feed with fresh request ID failed: %v", err)
	}
	if out.Balance != 80 {
		t.Fatalf("expected balance 80 after second distinct feed, got %d", out.Balance)
	}
}

// TestFeedConcurrentSameRequestReplaysOnce verifies concurrent delivery of one
// idempotency key produces one feeding row and returns the same result to both
// callers instead of leaking the backing unique-key conflict.
func TestFeedConcurrentSameRequestReplaysOnce(t *testing.T) {
	ctx := context.Background()
	setupPostgreSQLFeedingDB(t, ctx)
	svc := newFeedingServiceForTest(&fakeIronLocation{})
	insertQuoteRow(t, ctx, "stable quote", quoteEnabledOn)

	niuID := stageActiveNiu(t, ctx, "NIU-FEED-CONCURRENT", 30.0, 103.0)
	user := insertUserRow(t, ctx, "openid-feed-concurrent")
	creditGrass(t, ctx, user, 100)

	results := make(chan *FeedOutput, 2)
	errs := make(chan error, 2)
	var wg sync.WaitGroup
	for i := 0; i < 2; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			out, err := svc.Feed(ctx, user, &FeedInput{NiuId: niuID, BaseAmount: 10, RequestId: "req-feed-concurrent"})
			results <- out
			errs <- err
		}()
	}
	wg.Wait()
	close(results)
	close(errs)

	for err := range errs {
		if err != nil {
			t.Fatalf("concurrent replay failed: %v", err)
		}
	}
	var first *FeedOutput
	for out := range results {
		if first == nil {
			first = out
			continue
		}
		if !reflect.DeepEqual(first, out) {
			t.Fatalf("concurrent responses differ: first=%+v second=%+v", first, out)
		}
	}
	count, err := dao.Feeding.Ctx(ctx).Where(dao.Feeding.Columns().UserId, user).Count()
	if err != nil || count != 1 {
		t.Fatalf("expected one feeding row, count=%d err=%v", count, err)
	}
	if balance := feedingBalanceOf(t, ctx, user); balance != 90 {
		t.Fatalf("expected one deduction and balance 90, got %d", balance)
	}
}

// TestFeedConcurrentReplayPrecedesPositionFailure stages one request inside the
// player-locked transaction while a duplicate sees a position dependency error.
// Once the first request commits, the duplicate must replay it instead of
// surfacing the stale prefetch error.
func TestFeedConcurrentReplayPrecedesPositionFailure(t *testing.T) {
	ctx := context.Background()
	setupPostgreSQLFeedingDB(t, ctx)
	baseGrass := grasssvc.New(nil, grasssvc.Config{CheckinMinAmount: 1, CheckinMaxAmount: 1})
	blockingGrass := &blockingGrassService{
		Service: baseGrass,
		entered: make(chan struct{}),
		release: make(chan struct{}),
	}
	iron := &failAfterFirstIronLocation{
		err:        errors.New("iron positions unavailable"),
		secondCall: make(chan struct{}),
	}
	svc := New(blockingGrass, iron, nil, Config{IronBonusThresholdMeters: 12})
	niuID := stageActiveNiu(t, ctx, "NIU-FEED-DEPENDENCY-REPLAY", 30.0, 103.0)
	user := insertUserRow(t, ctx, "openid-feed-dependency-replay")
	creditGrass(t, ctx, user, 100)
	in := &FeedInput{NiuId: niuID, BaseAmount: 10, RequestId: "req-feed-dependency-replay"}

	type outcome struct {
		out *FeedOutput
		err error
	}
	firstResult := make(chan outcome, 1)
	secondResult := make(chan outcome, 1)
	go func() {
		out, err := svc.Feed(ctx, user, in)
		firstResult <- outcome{out: out, err: err}
	}()
	<-blockingGrass.entered
	go func() {
		out, err := svc.Feed(ctx, user, in)
		secondResult <- outcome{out: out, err: err}
	}()
	<-iron.secondCall
	close(blockingGrass.release)

	first := <-firstResult
	second := <-secondResult
	if first.err != nil || second.err != nil {
		t.Fatalf("concurrent stable replay failed: first=%v second=%v", first.err, second.err)
	}
	if !reflect.DeepEqual(first.out, second.out) {
		t.Fatalf("dependency failure changed replay: first=%+v second=%+v", first.out, second.out)
	}
}

// TestFeedRequiresRequestID verifies internal callers cannot bypass the public
// idempotency-key requirement.
func TestFeedRequiresRequestID(t *testing.T) {
	svc := newFeedingServiceForTest(&fakeIronLocation{})
	_, err := svc.Feed(context.Background(), 1, &FeedInput{NiuId: 1, BaseAmount: 10})
	assertBizCode(t, err, CodeRequestIDRequired.RuntimeCode())
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

	_, err := svc.Feed(ctx, user, &FeedInput{NiuId: niuID, BaseAmount: 10, RequestId: "req-feed-inactive"})
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

	_, err := svc.Feed(ctx, user, &FeedInput{NiuId: niuID, BaseAmount: 10, RequestId: "req-feed-poor"})
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
