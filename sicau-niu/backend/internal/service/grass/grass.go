// Package grass implements the sicau-niu C4 ledger grass-account capability:
// the per-player grass balance account, the append-only transaction ledger, the
// daily check-in grant, and the reusable in-transaction accounting primitive
// (ApplyDelta) consumed by feeding, steal and gift. Every grass change is written
// as a ledger transaction and applied to the balance in the same transaction, so
// the balance is always reconstructable from the ledger. Debits that would drive
// the balance negative are rejected. The account is created lazily on first
// access with a zero balance. Every player-facing read and the check-in are
// isolated to the authenticated player; the cross-capability ApplyDelta operates
// on an explicit user ID supplied by the caller and runs inside the caller's
// transaction so multi-party transfers stay atomic. All store access uses the
// generated DAO/DO objects so GoFrame manages soft-delete and timestamp columns.
package grass

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"

	rulessvc "lina-plugin-sicau-niu/backend/internal/service/rules"
)

// Config carries the plain-value runtime configuration for the grass capability.
// It holds only scalar tuning values and never runtime service dependencies.
type Config struct {
	// CheckinMinAmount is the inclusive lower bound of the random daily check-in grant.
	CheckinMinAmount int
	// CheckinMaxAmount is the inclusive upper bound of the random daily check-in grant.
	CheckinMaxAmount int
}

// Service defines the C4 grass account, ledger and check-in contract, plus the
// in-transaction accounting primitive reused by feeding, steal and gift.
type Service interface {
	// Checkin grants the authenticated player a random grass amount once per
	// natural day, writing a checkin ledger transaction and crediting the balance
	// in one transaction. It returns CodeAlreadyCheckedIn when the player already
	// checked in today, or a store bizerr on failure.
	Checkin(ctx context.Context, playerID int64) (out *CheckinResult, err error)
	// Account returns the player's current grass balance and a bounded page of
	// recent ledger transactions (newest first). It is isolated to the current
	// player and creates the account lazily (zero balance) when absent. It returns
	// a query bizerr on store failure.
	Account(ctx context.Context, playerID int64) (out *AccountView, err error)
	// ApplyDelta applies a signed grass change to userID inside the caller's open
	// transaction: it locks (or lazily creates) the account row, rejects a debit
	// that would make the balance negative with CodeInsufficientBalance, updates
	// the balance and appends a matching ledger transaction. The caller MUST invoke
	// it within a dao Transaction closure so the balance and ledger stay
	// consistent and multi-party transfers are atomic. It returns the resulting
	// balance, or a validation/store bizerr on failure.
	ApplyDelta(ctx context.Context, tx gdb.TX, userID int64, delta int64, txnType TxnType, refID int64) (newBalance int64, err error)
}

// Interface compliance assertion for the default grass service implementation.
var _ Service = (*serviceImpl)(nil)

// serviceImpl implements Service against the plugin-owned grass account and
// ledger tables. The optional rules service supplies operator-maintained
// check-in bounds; constructor scalar values remain the fallback.
type serviceImpl struct {
	rulesSvc   rulessvc.Service // rulesSvc supplies operator-maintained check-in bounds when injected.
	checkinMin int              // checkinMin is the inclusive lower bound of the daily check-in grant.
	checkinMax int              // checkinMax is the inclusive upper bound of the daily check-in grant.
}

// New creates a grass service with an optional runtime-rule service and fallback
// plain-value check-in configuration. The fallback range is normalized so min<=max
// and both are positive, keeping the random grant well-defined regardless of
// configuration order.
func New(rulesSvc rulessvc.Service, config Config) Service {
	min, max := normalizeCheckinRange(config.CheckinMinAmount, config.CheckinMaxAmount)
	return &serviceImpl{rulesSvc: rulesSvc, checkinMin: min, checkinMax: max}
}
