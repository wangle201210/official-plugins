// grass_checkin_test.go covers the DB-gated daily check-in: a first check-in
// grants and credits grass, a repeat check-in on the same day is rejected, and the
// granted amount stays within the configured range. The check-in range
// normalization is also covered as a pure-logic check.

package grass

import (
	"context"
	"errors"
	"reflect"
	"strings"
	"testing"

	"lina-plugin-sicau-niu/backend/internal/dao"
	rulessvc "lina-plugin-sicau-niu/backend/internal/service/rules"
)

type failingCheckinRules struct {
	rulessvc.Service
	err error
}

func (s *failingCheckinRules) CheckinRange(context.Context) (int, int, error) {
	return 0, 0, s.err
}

// TestCheckinGrantsAndCredits verifies a first check-in grants the configured
// amount, credits the ledger and keeps the balance equal to the ledger sum.
func TestCheckinGrantsAndCredits(t *testing.T) {
	ctx := context.Background()
	setupPostgreSQLGrassDB(t, ctx)
	svc := newGrassServiceForTest()
	user := insertUserRow(t, ctx, "openid-checkin")

	result, err := svc.Checkin(ctx, user, "checkin-grant")
	if err != nil {
		t.Fatalf("checkin failed: %v", err)
	}
	if result.Amount != 30 {
		t.Fatalf("expected fixed grant 30, got %d", result.Amount)
	}
	if result.Balance != 30 {
		t.Fatalf("expected balance 30, got %d", result.Balance)
	}
	if sum := ledgerSum(t, ctx, user); sum != result.Balance {
		t.Fatalf("balance %d does not equal ledger sum %d", result.Balance, sum)
	}
}

// TestCheckinRepeatRejected verifies a second check-in on the same day is rejected
// and does not credit additional grass.
func TestCheckinRepeatRejected(t *testing.T) {
	ctx := context.Background()
	setupPostgreSQLGrassDB(t, ctx)
	svc := newGrassServiceForTest()
	user := insertUserRow(t, ctx, "openid-checkin-repeat")

	if _, err := svc.Checkin(ctx, user, "checkin-repeat-first"); err != nil {
		t.Fatalf("first checkin failed: %v", err)
	}
	_, err := svc.Checkin(ctx, user, "checkin-repeat-second")
	assertBizCode(t, err, CodeAlreadyCheckedIn.RuntimeCode())

	account, err := svc.Account(ctx, user)
	if err != nil {
		t.Fatalf("account read failed: %v", err)
	}
	if account.Balance != 30 {
		t.Fatalf("expected balance unchanged at 30 after rejected repeat, got %d", account.Balance)
	}
}

// TestCheckinRequestReplayReturnsExactResult verifies retrying the same mandatory
// request ID returns the persisted first response without another grant or ledger
// transaction.
func TestCheckinRequestReplayReturnsExactResult(t *testing.T) {
	ctx := context.Background()
	setupPostgreSQLGrassDB(t, ctx)
	svc := newGrassServiceForTest()
	user := insertUserRow(t, ctx, "openid-checkin-replay")

	first, err := svc.Checkin(ctx, user, "checkin-replay")
	if err != nil {
		t.Fatalf("first checkin failed: %v", err)
	}
	replayed, err := svc.Checkin(ctx, user, "checkin-replay")
	if err != nil {
		t.Fatalf("checkin replay failed: %v", err)
	}
	if !reflect.DeepEqual(first, replayed) {
		t.Fatalf("checkin replay changed response: first=%+v replay=%+v", first, replayed)
	}
	checkinCount, err := dao.Checkin.Ctx(ctx).Where(dao.Checkin.Columns().UserId, user).Count()
	if err != nil {
		t.Fatalf("count checkins failed: %v", err)
	}
	if checkinCount != 1 {
		t.Fatalf("expected one checkin row after replay, got %d", checkinCount)
	}
	txnCount, err := dao.GrassTxn.Ctx(ctx).Where(dao.GrassTxn.Columns().UserId, user).Count()
	if err != nil {
		t.Fatalf("count grass transactions failed: %v", err)
	}
	if txnCount != 1 {
		t.Fatalf("expected one ledger transaction after replay, got %d", txnCount)
	}
}

// TestCheckinReplayPrecedesRuleFailure verifies a completed request remains
// replayable when the live rule source later becomes unavailable.
func TestCheckinReplayPrecedesRuleFailure(t *testing.T) {
	ctx := context.Background()
	setupPostgreSQLGrassDB(t, ctx)
	svc := newGrassServiceForTest().(*serviceImpl)
	user := insertUserRow(t, ctx, "openid-checkin-rule-failure")

	first, err := svc.Checkin(ctx, user, "checkin-rule-failure")
	if err != nil {
		t.Fatalf("first checkin failed: %v", err)
	}
	svc.rulesSvc = &failingCheckinRules{err: errors.New("rule source unavailable")}
	replayed, err := svc.Checkin(ctx, user, "checkin-rule-failure")
	if err != nil {
		t.Fatalf("completed checkin should replay before reading rules: %v", err)
	}
	if !reflect.DeepEqual(first, replayed) {
		t.Fatalf("rule failure changed replay: first=%+v replay=%+v", first, replayed)
	}
}

// TestCheckinRequiresRequestID verifies internal callers cannot bypass the
// mandatory idempotency-key contract enforced by the HTTP DTO.
func TestCheckinRequiresRequestID(t *testing.T) {
	svc := newGrassServiceForTest()
	for _, requestID := range []string{"", "   ", strings.Repeat("r", 65)} {
		_, err := svc.Checkin(context.Background(), 1, requestID)
		assertBizCode(t, err, CodeRequestIDRequired.RuntimeCode())
	}
}

// TestNormalizeCheckinRange verifies the range normalization keeps both bounds
// positive and ordered regardless of configuration order, without a database.
func TestNormalizeCheckinRange(t *testing.T) {
	cases := []struct {
		name             string
		inMin, inMax     int
		wantMin, wantMax int
	}{
		{"normal", 20, 50, 20, 50},
		{"reversed", 50, 20, 50, 50},
		{"non-positive-min", 0, 40, 1, 40},
		{"both-non-positive", -5, -1, 1, 1},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			gotMin, gotMax := normalizeCheckinRange(tc.inMin, tc.inMax)
			if gotMin != tc.wantMin || gotMax != tc.wantMax {
				t.Fatalf("normalizeCheckinRange(%d,%d) = (%d,%d), want (%d,%d)",
					tc.inMin, tc.inMax, gotMin, gotMax, tc.wantMin, tc.wantMax)
			}
		})
	}
}
