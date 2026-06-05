// grass_checkin.go implements the daily check-in grant. A check-in inserts a
// uniquely-keyed (user_id, checkin_date) row, grants a random amount within the
// configured range and credits it to the player's ledger account, all inside one
// transaction so the check-in record, the credit and the balance stay consistent.

package grass

import (
	"context"
	"time"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/util/grand"

	"lina-core/pkg/bizerr"
	"lina-plugin-sicau-niu/backend/internal/dao"
	"lina-plugin-sicau-niu/backend/internal/model/do"
)

// checkinDateLayout is the YYYY-MM-DD natural-day key used for the per-day
// check-in uniqueness check.
const checkinDateLayout = "2006-01-02"

// CheckinResult is the outcome of a successful daily check-in.
type CheckinResult struct {
	// Amount is the grass amount granted by this check-in.
	Amount int
	// Balance is the player's grass balance after the check-in credit.
	Balance int64
}

// Checkin grants the player a random daily check-in amount once per natural day.
func (s *serviceImpl) Checkin(ctx context.Context, playerID int64) (*CheckinResult, error) {
	if playerID <= 0 {
		return nil, bizerr.NewCode(CodeQueryFailed)
	}

	today := time.Now().Format(checkinDateLayout)
	checkinMin, checkinMax, err := s.checkinRange(ctx)
	if err != nil {
		return nil, err
	}
	amount := grand.N(checkinMin, checkinMax)

	var result *CheckinResult
	err = dao.Checkin.Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {
		alreadyCheckedIn, txErr := dao.Checkin.Ctx(ctx).
			Where(dao.Checkin.Columns().UserId, playerID).
			Where(dao.Checkin.Columns().CheckinDate, today).
			Count()
		if txErr != nil {
			return bizerr.WrapCode(txErr, CodeQueryFailed)
		}
		if alreadyCheckedIn > 0 {
			return bizerr.NewCode(CodeAlreadyCheckedIn)
		}

		checkinID, txErr := dao.Checkin.Ctx(ctx).Data(do.Checkin{
			UserId:      playerID,
			CheckinDate: today,
			Amount:      amount,
		}).InsertAndGetId()
		if txErr != nil {
			return bizerr.WrapCode(txErr, CodeWriteFailed)
		}

		newBalance, applyErr := s.ApplyDelta(ctx, tx, playerID, int64(amount), TxnTypeCheckin, checkinID)
		if applyErr != nil {
			return applyErr
		}

		result = &CheckinResult{Amount: amount, Balance: newBalance}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return result, nil
}

// checkinRange returns the operator-maintained check-in range when rules are
// injected, otherwise the constructor fallback.
func (s *serviceImpl) checkinRange(ctx context.Context) (int, int, error) {
	if s.rulesSvc == nil {
		return s.checkinMin, s.checkinMax, nil
	}
	return s.rulesSvc.CheckinRange(ctx)
}

// normalizeCheckinRange clamps the configured check-in grant range so the random
// grant is always well-defined: both bounds become at least 1 and the lower
// bound never exceeds the upper bound regardless of configuration order.
func normalizeCheckinRange(min, max int) (int, int) {
	if min < 1 {
		min = 1
	}
	if max < min {
		max = min
	}
	return min, max
}
