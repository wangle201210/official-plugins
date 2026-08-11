// grass_ledger.go implements the ledger account view and the in-transaction
// accounting primitive ApplyDelta. ApplyDelta locks (or lazily creates) the
// account row inside the caller's transaction, enforces the non-negative balance
// invariant on debits, updates the balance and appends a matching ledger
// transaction so the balance always equals the signed sum of the ledger.

package grass

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"

	"lina-core/pkg/apitime"
	"lina-core/pkg/bizerr"
	"lina-plugin-sicau-niu/backend/internal/activityday"
	"lina-plugin-sicau-niu/backend/internal/dao"
	"lina-plugin-sicau-niu/backend/internal/model/do"
	entitymodel "lina-plugin-sicau-niu/backend/internal/model/entity"
)

// recentTxnLimit bounds the recent-transaction page returned with the account
// view so the player account read never loads the full ledger.
const recentTxnLimit = 20

// AccountView is the player grass account projection: current balance plus a
// bounded page of recent ledger transactions (newest first).
type AccountView struct {
	// Balance is the player's current grass balance.
	Balance int64
	// Level is one plus each completed 100 effective feeding points.
	Level int
	// Exp is cumulative effective feeding experience.
	Exp int64
	// CheckedToday reports whether today's Beijing-time check-in exists.
	CheckedToday bool
	// Recent holds the most recent ledger transactions, newest first.
	Recent []*TxnView
}

// ProgressView contains the shared profile and grass-account progress fields.
type ProgressView struct {
	Level        int
	Exp          int64
	CheckedToday bool
}

// TxnView is one ledger transaction projected for the player account view.
type TxnView struct {
	// TxnType is the transaction type string.
	TxnType string
	// Delta is the signed grass change.
	Delta int64
	// CreatedAt is the transaction time as a Unix timestamp in milliseconds.
	CreatedAt *int64
}

// Account returns the player's current balance and recent ledger transactions.
func (s *serviceImpl) Account(ctx context.Context, playerID int64) (*AccountView, error) {
	if playerID <= 0 {
		return nil, bizerr.NewCode(CodeQueryFailed)
	}

	balance, err := s.readBalance(ctx, playerID)
	if err != nil {
		return nil, err
	}

	rows := make([]*entitymodel.GrassTxn, 0, recentTxnLimit)
	err = dao.GrassTxn.Ctx(ctx).
		Fields(
			dao.GrassTxn.Columns().TxnType,
			dao.GrassTxn.Columns().Delta,
			dao.GrassTxn.Columns().CreatedAt,
		).
		Where(dao.GrassTxn.Columns().UserId, playerID).
		OrderDesc(dao.GrassTxn.Columns().Id).
		Limit(recentTxnLimit).
		Scan(&rows)
	if err != nil {
		return nil, bizerr.WrapCode(err, CodeQueryFailed)
	}

	recent := make([]*TxnView, 0, len(rows))
	for _, row := range rows {
		recent = append(recent, &TxnView{
			TxnType:   row.TxnType,
			Delta:     row.Delta,
			CreatedAt: apitime.Milli(row.CreatedAt),
		})
	}
	progress, err := s.Progress(ctx, playerID)
	if err != nil {
		return nil, err
	}
	return &AccountView{Balance: balance, Level: progress.Level, Exp: progress.Exp, CheckedToday: progress.CheckedToday, Recent: recent}, nil
}

// Progress computes level and check-in state with two fixed aggregate queries.
func (s *serviceImpl) Progress(ctx context.Context, playerID int64) (*ProgressView, error) {
	if playerID <= 0 {
		return nil, bizerr.NewCode(CodeQueryFailed)
	}
	value, err := dao.Feeding.Ctx(ctx).Where(dao.Feeding.Columns().UserId, playerID).Sum(dao.Feeding.Columns().EffectAmount)
	if err != nil {
		return nil, bizerr.WrapCode(err, CodeQueryFailed)
	}
	exp := int64(value)
	checked, err := dao.Checkin.Ctx(ctx).Where(do.Checkin{UserId: playerID, CheckinDate: activityday.Today()}).Count()
	if err != nil {
		return nil, bizerr.WrapCode(err, CodeQueryFailed)
	}
	return &ProgressView{Level: 1 + int(exp/100), Exp: exp, CheckedToday: checked > 0}, nil
}

// readBalance returns the player's current balance, reading only the balance
// column. A missing account row reports a zero balance without creating it.
func (s *serviceImpl) readBalance(ctx context.Context, userID int64) (int64, error) {
	balance, err := dao.GrassAccount.Ctx(ctx).
		Fields(dao.GrassAccount.Columns().Balance).
		Where(dao.GrassAccount.Columns().UserId, userID).
		Value()
	if err != nil {
		return 0, bizerr.WrapCode(err, CodeQueryFailed)
	}
	return balance.Int64(), nil
}

// ApplyDelta applies a signed grass change to userID inside the caller's
// transaction and appends the matching ledger transaction. The tx parameter
// documents the in-transaction contract; the locked read and writes run on ctx,
// which carries the transaction inside a dao Transaction closure.
func (s *serviceImpl) ApplyDelta(ctx context.Context, tx gdb.TX, userID int64, delta int64, txnType TxnType, refID int64) (int64, error) {
	if userID <= 0 {
		return 0, bizerr.NewCode(CodeQueryFailed)
	}
	if delta == 0 {
		return 0, bizerr.NewCode(CodeInvalidDelta)
	}
	if tx == nil {
		return 0, bizerr.NewCode(CodeWriteFailed)
	}

	balance, err := s.lockOrCreateBalance(ctx, userID)
	if err != nil {
		return 0, err
	}

	newBalance := balance + delta
	if newBalance < 0 {
		return 0, bizerr.NewCode(CodeInsufficientBalance)
	}

	if _, err = dao.GrassAccount.Ctx(ctx).
		Where(dao.GrassAccount.Columns().UserId, userID).
		Data(do.GrassAccount{Balance: newBalance}).
		Update(); err != nil {
		return 0, bizerr.WrapCode(err, CodeWriteFailed)
	}

	if _, err = dao.GrassTxn.Ctx(ctx).Data(do.GrassTxn{
		UserId:  userID,
		Delta:   delta,
		TxnType: txnType.String(),
		RefId:   refID,
	}).Insert(); err != nil {
		return 0, bizerr.WrapCode(err, CodeWriteFailed)
	}
	return newBalance, nil
}

// lockOrCreateBalance loads the account balance under a row lock, lazily creating
// a zero-balance account when the player has none yet. It must run inside the
// caller's transaction so the lock serializes concurrent debits/credits.
func (s *serviceImpl) lockOrCreateBalance(ctx context.Context, userID int64) (int64, error) {
	var account *entitymodel.GrassAccount
	err := dao.GrassAccount.Ctx(ctx).
		Where(dao.GrassAccount.Columns().UserId, userID).
		LockUpdate().
		Scan(&account)
	if err != nil {
		return 0, bizerr.WrapCode(err, CodeQueryFailed)
	}
	if account != nil {
		return account.Balance, nil
	}

	if _, err = dao.GrassAccount.Ctx(ctx).Data(do.GrassAccount{
		UserId:  userID,
		Balance: 0,
	}).Insert(); err != nil {
		return 0, bizerr.WrapCode(err, CodeWriteFailed)
	}
	return 0, nil
}
