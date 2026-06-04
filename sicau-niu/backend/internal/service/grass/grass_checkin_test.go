// grass_checkin_test.go covers the DB-gated daily check-in: a first check-in
// grants and credits grass, a repeat check-in on the same day is rejected, and the
// granted amount stays within the configured range. The check-in range
// normalization is also covered as a pure-logic check.

package grass

import (
	"context"
	"testing"
)

// TestCheckinGrantsAndCredits verifies a first check-in grants the configured
// amount, credits the ledger and keeps the balance equal to the ledger sum.
func TestCheckinGrantsAndCredits(t *testing.T) {
	ctx := context.Background()
	setupPostgreSQLGrassDB(t, ctx)
	svc := newGrassServiceForTest()
	user := insertUserRow(t, ctx, "openid-checkin")

	result, err := svc.Checkin(ctx, user)
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

	if _, err := svc.Checkin(ctx, user); err != nil {
		t.Fatalf("first checkin failed: %v", err)
	}
	_, err := svc.Checkin(ctx, user)
	assertBizCode(t, err, CodeAlreadyCheckedIn.RuntimeCode())

	account, err := svc.Account(ctx, user)
	if err != nil {
		t.Fatalf("account read failed: %v", err)
	}
	if account.Balance != 30 {
		t.Fatalf("expected balance unchanged at 30 after rejected repeat, got %d", account.Balance)
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
