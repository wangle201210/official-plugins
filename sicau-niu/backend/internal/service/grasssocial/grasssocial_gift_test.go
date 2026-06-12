// grasssocial_gift_test.go covers the DB-gated gift action: a gift moves grass
// between both players in one transaction and notifies the recipient, an amount
// below the per-gift minimum is rejected, an insufficient balance is rejected, and
// the daily gift limit is enforced.

package grasssocial

import (
	"context"
	"testing"

	"lina-plugin-sicau-niu/backend/internal/dao"
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

	out, err := svc.Gift(ctx, giver, &GiftInput{ToUserId: recipient, Amount: 12})
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

// TestGiftDuplicateRequestRejected verifies a retry carrying an already-recorded
// request ID is rejected with CodeDuplicateRequest and moves no further grass.
func TestGiftDuplicateRequestRejected(t *testing.T) {
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

	_, err = svc.Gift(ctx, giver, &GiftInput{ToUserId: recipient, Amount: 12, RequestId: "req-gift-1"})
	assertBizCode(t, err, CodeDuplicateRequest.RuntimeCode())

	if got := balanceOf(t, ctx, giver); got != 88 {
		t.Fatalf("expected giver balance unchanged at 88 after duplicate, got %d", got)
	}
	if got := balanceOf(t, ctx, recipient); got != 12 {
		t.Fatalf("expected recipient balance unchanged at 12 after duplicate, got %d", got)
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

	_, err := svc.Gift(ctx, giver, &GiftInput{ToUserId: recipient, Amount: 11})
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

	_, err := svc.Gift(ctx, giver, &GiftInput{ToUserId: recipient, Amount: 12})
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
		if _, err := svc.Gift(ctx, giver, &GiftInput{ToUserId: recipient, Amount: 12}); err != nil {
			t.Fatalf("gift %d failed: %v", i, err)
		}
	}
	_, err := svc.Gift(ctx, giver, &GiftInput{ToUserId: recipient, Amount: 12})
	assertBizCode(t, err, CodeGiftLimitReached.RuntimeCode())
}
