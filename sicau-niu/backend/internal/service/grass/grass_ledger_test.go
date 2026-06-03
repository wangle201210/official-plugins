// grass_ledger_test.go covers the DB-gated ledger primitive: the balance always
// equals the signed ledger sum, a debit beyond the balance is rejected, the
// account is created lazily, and zero/no-tx deltas are rejected. Pure-logic checks
// run without a database.

package grass

import (
	"context"
	"testing"

	"github.com/gogf/gf/v2/database/gdb"

	"lina-plugin-sicau-niu/backend/internal/dao"
)

// applyDeltaTx runs one ApplyDelta inside its own transaction and returns the
// resulting balance, mirroring how feeding/steal/gift call it within a tx.
func applyDeltaTx(ctx context.Context, svc Service, userID int64, delta int64, txnType TxnType, refID int64) (int64, error) {
	var balance int64
	err := dao.GrassAccount.Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {
		applied, applyErr := svc.ApplyDelta(ctx, tx, userID, delta, txnType, refID)
		if applyErr != nil {
			return applyErr
		}
		balance = applied
		return nil
	})
	return balance, err
}

// TestApplyDeltaBalanceEqualsLedger verifies the balance equals the signed sum of
// the ledger after a sequence of credits and debits.
func TestApplyDeltaBalanceEqualsLedger(t *testing.T) {
	ctx := context.Background()
	setupPostgreSQLGrassDB(t, ctx)
	svc := newGrassServiceForTest()
	user := insertUserRow(t, ctx, "openid-ledger")

	if _, err := applyDeltaTx(ctx, svc, user, 100, TxnTypeCheckin, 0); err != nil {
		t.Fatalf("credit failed: %v", err)
	}
	if _, err := applyDeltaTx(ctx, svc, user, -30, TxnTypeFeed, 0); err != nil {
		t.Fatalf("debit failed: %v", err)
	}
	balance, err := applyDeltaTx(ctx, svc, user, 12, TxnTypeGiftIn, 0)
	if err != nil {
		t.Fatalf("credit failed: %v", err)
	}

	if balance != 82 {
		t.Fatalf("expected balance 82, got %d", balance)
	}
	if sum := ledgerSum(t, ctx, user); sum != balance {
		t.Fatalf("balance %d does not equal ledger sum %d", balance, sum)
	}
}

// TestApplyDeltaInsufficientBalanceRejected verifies a debit beyond the balance is
// rejected and leaves the balance and ledger unchanged.
func TestApplyDeltaInsufficientBalanceRejected(t *testing.T) {
	ctx := context.Background()
	setupPostgreSQLGrassDB(t, ctx)
	svc := newGrassServiceForTest()
	user := insertUserRow(t, ctx, "openid-insufficient")

	if _, err := applyDeltaTx(ctx, svc, user, 20, TxnTypeCheckin, 0); err != nil {
		t.Fatalf("credit failed: %v", err)
	}
	_, err := applyDeltaTx(ctx, svc, user, -50, TxnTypeFeed, 0)
	assertBizCode(t, err, CodeInsufficientBalance.RuntimeCode())

	account, err := svc.Account(ctx, user)
	if err != nil {
		t.Fatalf("account read failed: %v", err)
	}
	if account.Balance != 20 {
		t.Fatalf("expected balance unchanged at 20, got %d", account.Balance)
	}
	if sum := ledgerSum(t, ctx, user); sum != 20 {
		t.Fatalf("expected ledger sum unchanged at 20, got %d", sum)
	}
}

// TestAccountLazyCreation verifies a never-credited player reads a zero balance
// and an empty recent list without error.
func TestAccountLazyCreation(t *testing.T) {
	ctx := context.Background()
	setupPostgreSQLGrassDB(t, ctx)
	svc := newGrassServiceForTest()
	user := insertUserRow(t, ctx, "openid-lazy")

	account, err := svc.Account(ctx, user)
	if err != nil {
		t.Fatalf("account read failed: %v", err)
	}
	if account.Balance != 0 || len(account.Recent) != 0 {
		t.Fatalf("expected zero balance and empty recent, got %+v", account)
	}
}

// TestApplyDeltaZeroRejected verifies a zero delta is rejected as invalid without
// a database.
func TestApplyDeltaZeroRejected(t *testing.T) {
	ctx := context.Background()
	svc := newGrassServiceForTest()
	// A nil tx with a zero delta still rejects on the delta validation that runs
	// before any store access, so this stays a pure-logic check.
	_, err := svc.ApplyDelta(ctx, nil, 1, 0, TxnTypeCheckin, 0)
	assertBizCode(t, err, CodeInvalidDelta.RuntimeCode())
}
