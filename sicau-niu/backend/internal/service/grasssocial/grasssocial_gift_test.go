// grasssocial_gift_test.go covers the DB-gated gift action: a gift moves grass
// between both players in one transaction and notifies the recipient, an amount
// below the per-gift minimum is rejected, an insufficient balance is rejected, and
// the daily gift limit is enforced.

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
	rulessvc "lina-plugin-sicau-niu/backend/internal/service/rules"
)

// TestGiftMovesGrassAndNotifies verifies a gift debits the giver, credits the
// recipient in one transaction and writes a gift-received message.
func TestGiftMovesGrassAndNotifies(t *testing.T) {
	ctx := context.Background()
	setupPostgreSQLSocialDB(t, ctx)
	svc := newSocialServiceForTest()

	giver := insertUserRow(t, ctx, "openid-gift-giver")
	recipient := insertUserRow(t, ctx, "openid-gift-recipient")
	seedGrass(t, ctx, giver, 100)

	out, err := svc.Gift(ctx, giver, &GiftInput{ToUserId: recipient, Amount: 12, RequestId: "req-gift-normal"})
	if err != nil {
		t.Fatalf("gift failed: %v", err)
	}
	if out.Balance != 88 {
		t.Fatalf("expected giver balance 88, got %d", out.Balance)
	}
	if got := balanceOf(t, ctx, recipient); got != 12 {
		t.Fatalf("expected recipient balance 12, got %d", got)
	}
	if n := inboxCount(t, ctx, recipient, MsgTypeGiftReceived); n != 1 {
		t.Fatalf("expected 1 gift-received notification, got %d", n)
	}
}

// TestGiftUsesOneLockedRuleSnapshot verifies one new gift reads live social
// rules exactly once after locking and uses that snapshot for both minimum and
// daily-limit checks. A later rule version is observed only by the next write.
func TestGiftUsesOneLockedRuleSnapshot(t *testing.T) {
	ctx := context.Background()
	setupPostgreSQLSocialDB(t, ctx)
	rules := newSequenceRulesService(
		rulessvc.SocialRules{GiftDailyLimit: 2, GiftMinAmount: 12},
		rulessvc.SocialRules{GiftDailyLimit: 1, GiftMinAmount: 99},
	)
	svc := newSocialServiceWithRules(rules)

	giver := insertUserRow(t, ctx, "openid-gift-snapshot-giver")
	recipient := insertUserRow(t, ctx, "openid-gift-snapshot-recipient")
	seedGrass(t, ctx, giver, 100)
	if _, err := dao.Gift.Ctx(ctx).Data(do.Gift{
		FromUserId: giver,
		ToUserId:   recipient,
		Amount:     12,
		GiftDate:   activityday.Today(),
		RequestId:  "req-gift-snapshot-prior",
	}).Insert(); err != nil {
		t.Fatalf("insert prior gift failed: %v", err)
	}

	first, err := svc.Gift(ctx, giver, &GiftInput{ToUserId: recipient, Amount: 12, RequestId: "req-gift-snapshot"})
	if err != nil {
		t.Fatalf("gift with first rule snapshot failed: %v", err)
	}
	replayed, err := svc.Gift(ctx, giver, &GiftInput{ToUserId: recipient, Amount: 12, RequestId: "req-gift-snapshot"})
	if err != nil || !reflect.DeepEqual(first, replayed) {
		t.Fatalf("gift replay changed the first result: first=%+v replay=%+v err=%v", first, replayed, err)
	}
	if got := rules.callCount(); got != 1 {
		t.Fatalf("expected one rule read for gift plus replay, got %d", got)
	}

	_, err = svc.Gift(ctx, giver, &GiftInput{ToUserId: recipient, Amount: 12, RequestId: "req-gift-snapshot-next"})
	assertBizCode(t, err, CodeGiftAmountTooSmall.RuntimeCode())
	if got := rules.callCount(); got != 2 {
		t.Fatalf("expected the next gift to read the next rule snapshot once, got %d total reads", got)
	}
}

// TestGiftDuplicateRequestReplaysExactResult verifies a retry carrying an
// already-recorded request ID returns the first result and moves no further grass.
func TestGiftDuplicateRequestReplaysExactResult(t *testing.T) {
	ctx := context.Background()
	setupPostgreSQLSocialDB(t, ctx)
	svc := newSocialServiceForTest()

	giver := insertUserRow(t, ctx, "openid-gift-dup-giver")
	recipient := insertUserRow(t, ctx, "openid-gift-dup-recipient")
	seedGrass(t, ctx, giver, 100)

	out, err := svc.Gift(ctx, giver, &GiftInput{ToUserId: recipient, Amount: 12, RequestId: "req-gift-1"})
	if err != nil {
		t.Fatalf("first gift failed: %v", err)
	}
	if out.Balance != 88 {
		t.Fatalf("expected giver balance 88 after first gift, got %d", out.Balance)
	}

	replayed, err := svc.Gift(ctx, giver, &GiftInput{ToUserId: recipient, Amount: 12, RequestId: "req-gift-1"})
	if err != nil {
		t.Fatalf("gift replay failed: %v", err)
	}
	if !reflect.DeepEqual(out, replayed) {
		t.Fatalf("expected exact first response on replay, first=%+v replay=%+v", out, replayed)
	}

	if got := balanceOf(t, ctx, giver); got != 88 {
		t.Fatalf("expected giver balance unchanged at 88 after duplicate, got %d", got)
	}
	if got := balanceOf(t, ctx, recipient); got != 12 {
		t.Fatalf("expected recipient balance unchanged at 12 after duplicate, got %d", got)
	}
}

// TestGiftConcurrentDailyLimitEnforced verifies distinct concurrent request IDs
// from one giver cannot both pass a one-action daily cap.
func TestGiftConcurrentDailyLimitEnforced(t *testing.T) {
	ctx := context.Background()
	setupPostgreSQLSocialDB(t, ctx)
	svc := newSocialServiceWithLimits(2, 1)

	giver := insertUserRow(t, ctx, "openid-gift-concurrent-giver")
	recipient := insertUserRow(t, ctx, "openid-gift-concurrent-recipient")
	seedGrass(t, ctx, giver, 100)

	errs := make(chan error, 2)
	var wg sync.WaitGroup
	for _, requestID := range []string{"req-gift-a", "req-gift-b"} {
		wg.Add(1)
		go func(requestID string) {
			defer wg.Done()
			_, err := svc.Gift(ctx, giver, &GiftInput{ToUserId: recipient, Amount: 12, RequestId: requestID})
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
		if isBizCode(err, CodeGiftLimitReached) {
			limitErrors++
			continue
		}
		t.Fatalf("unexpected concurrent gift error: %v", err)
	}
	if successes != 1 || limitErrors != 1 {
		t.Fatalf("expected one success and one limit error, successes=%d limits=%d", successes, limitErrors)
	}
	count, err := dao.Gift.Ctx(ctx).Where(dao.Gift.Columns().FromUserId, giver).Count()
	if err != nil || count != 1 {
		t.Fatalf("expected one gift row, count=%d err=%v", count, err)
	}
}

// TestGiftConcurrentSameRequestReplaysOnce verifies concurrent delivery of one
// idempotency key returns one stable outcome without a duplicate-key error.
func TestGiftConcurrentSameRequestReplaysOnce(t *testing.T) {
	ctx := context.Background()
	setupPostgreSQLSocialDB(t, ctx)
	svc := newSocialServiceForTest()

	giver := insertUserRow(t, ctx, "openid-gift-replay-giver")
	recipient := insertUserRow(t, ctx, "openid-gift-replay-recipient")
	seedGrass(t, ctx, giver, 100)

	results := make(chan *GiftResult, 2)
	errs := make(chan error, 2)
	var wg sync.WaitGroup
	for i := 0; i < 2; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			out, err := svc.Gift(ctx, giver, &GiftInput{ToUserId: recipient, Amount: 12, RequestId: "req-gift-same"})
			results <- out
			errs <- err
		}()
	}
	wg.Wait()
	close(results)
	close(errs)
	for err := range errs {
		if err != nil {
			t.Fatalf("concurrent gift replay failed: %v", err)
		}
	}
	var first *GiftResult
	for out := range results {
		if first == nil {
			first = out
			continue
		}
		if !reflect.DeepEqual(first, out) {
			t.Fatalf("concurrent gift responses differ: first=%+v second=%+v", first, out)
		}
	}
	count, err := dao.Gift.Ctx(ctx).Where(dao.Gift.Columns().FromUserId, giver).Count()
	if err != nil || count != 1 {
		t.Fatalf("expected one gift row, count=%d err=%v", count, err)
	}
}

// TestGiftReciprocalTransfersUseStableLockOrder verifies two players gifting
// each other concurrently do not deadlock by locking user rows in opposite order.
func TestGiftReciprocalTransfersUseStableLockOrder(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	setupPostgreSQLSocialDB(t, ctx)
	svc := newSocialServiceForTest()
	first := insertUserRow(t, ctx, "openid-gift-reciprocal-a")
	second := insertUserRow(t, ctx, "openid-gift-reciprocal-b")
	seedGrass(t, ctx, first, 100)
	seedGrass(t, ctx, second, 100)

	errCh := make(chan error, 2)
	var wg sync.WaitGroup
	for _, call := range []struct {
		from, to int64
		request  string
	}{{first, second, "req-gift-a-to-b"}, {second, first, "req-gift-b-to-a"}} {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, err := svc.Gift(ctx, call.from, &GiftInput{ToUserId: call.to, Amount: 12, RequestId: call.request})
			errCh <- err
		}()
	}
	wg.Wait()
	close(errCh)
	for err := range errCh {
		if err != nil {
			t.Fatalf("reciprocal gift failed: %v", err)
		}
	}
	if got := balanceOf(t, ctx, first); got != 100 {
		t.Fatalf("expected first balance 100 after reciprocal gifts, got %d", got)
	}
	if got := balanceOf(t, ctx, second); got != 100 {
		t.Fatalf("expected second balance 100 after reciprocal gifts, got %d", got)
	}
}

// TestGiftBelowMinimumRejected verifies an amount below the per-gift minimum is
// rejected and moves no grass.
func TestGiftBelowMinimumRejected(t *testing.T) {
	ctx := context.Background()
	setupPostgreSQLSocialDB(t, ctx)
	svc := newSocialServiceForTest()

	giver := insertUserRow(t, ctx, "openid-gift-small-giver")
	recipient := insertUserRow(t, ctx, "openid-gift-small-recipient")
	seedGrass(t, ctx, giver, 100)

	_, err := svc.Gift(ctx, giver, &GiftInput{ToUserId: recipient, Amount: 11, RequestId: "req-gift-small"})
	assertBizCode(t, err, CodeGiftAmountTooSmall.RuntimeCode())

	if got := balanceOf(t, ctx, giver); got != 100 {
		t.Fatalf("expected giver balance unchanged at 100, got %d", got)
	}
}

// TestGiftInsufficientBalanceRejected verifies a gift beyond the balance is
// rejected and rolls back so no gift row remains.
func TestGiftInsufficientBalanceRejected(t *testing.T) {
	ctx := context.Background()
	setupPostgreSQLSocialDB(t, ctx)
	svc := newSocialServiceForTest()

	giver := insertUserRow(t, ctx, "openid-gift-poor-giver")
	recipient := insertUserRow(t, ctx, "openid-gift-poor-recipient")
	seedGrass(t, ctx, giver, 10)

	_, err := svc.Gift(ctx, giver, &GiftInput{ToUserId: recipient, Amount: 12, RequestId: "req-gift-poor"})
	if err == nil {
		t.Fatalf("expected insufficient-balance rejection, got nil")
	}

	count, err := dao.Gift.Ctx(ctx).Where(dao.Gift.Columns().FromUserId, giver).Count()
	if err != nil {
		t.Fatalf("count gift failed: %v", err)
	}
	if count != 0 {
		t.Fatalf("expected gift rolled back, got %d rows", count)
	}
}

// TestGiftDailyLimitEnforced verifies the giver cannot gift more than the daily
// limit. The fixture limit is 2.
func TestGiftDailyLimitEnforced(t *testing.T) {
	ctx := context.Background()
	setupPostgreSQLSocialDB(t, ctx)
	svc := newSocialServiceForTest()

	giver := insertUserRow(t, ctx, "openid-gift-limit-giver")
	recipient := insertUserRow(t, ctx, "openid-gift-limit-recipient")
	seedGrass(t, ctx, giver, 100)

	for i := 0; i < 2; i++ {
		if _, err := svc.Gift(ctx, giver, &GiftInput{ToUserId: recipient, Amount: 12, RequestId: fmt.Sprintf("req-gift-limit-%d", i)}); err != nil {
			t.Fatalf("gift %d failed: %v", i, err)
		}
	}
	_, err := svc.Gift(ctx, giver, &GiftInput{ToUserId: recipient, Amount: 12, RequestId: "req-gift-limit-over"})
	assertBizCode(t, err, CodeGiftLimitReached.RuntimeCode())
}

// TestGiftRequiresRequestID verifies internal callers cannot bypass the public
// idempotency-key requirement.
func TestGiftRequiresRequestID(t *testing.T) {
	svc := newSocialServiceForTest()
	_, err := svc.Gift(context.Background(), 1, &GiftInput{ToUserId: 2, Amount: 12})
	assertBizCode(t, err, CodeRequestIDRequired.RuntimeCode())
}
