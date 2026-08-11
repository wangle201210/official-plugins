// grass_checkin.go implements the daily check-in grant. A check-in inserts a
// uniquely-keyed (user_id, checkin_date) row keyed by the Beijing-time natural
// day, grants a random amount within the configured range and credits it to the
// player's ledger account, all inside one transaction so the check-in record,
// the credit and the balance stay consistent.

package grass

import (
	"context"
	"strings"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/util/grand"

	"lina-core/pkg/bizerr"
	"lina-plugin-sicau-niu/backend/internal/activityday"
	"lina-plugin-sicau-niu/backend/internal/dao"
	"lina-plugin-sicau-niu/backend/internal/model/do"
	entitymodel "lina-plugin-sicau-niu/backend/internal/model/entity"
)

// CheckinResult is the outcome of a successful daily check-in.
type CheckinResult struct {
	// Amount is the grass amount granted by this check-in.
	Amount int
	// Balance is the player's grass balance after the check-in credit.
	Balance int64
}

// Checkin grants the player a random daily check-in amount once per natural day.
func (s *serviceImpl) Checkin(ctx context.Context, playerID int64, requestIDs ...string) (*CheckinResult, error) {
	if playerID <= 0 {
		return nil, bizerr.NewCode(CodeQueryFailed)
	}

	today := activityday.Today()
	requestID := ""
	if len(requestIDs) > 0 {
		requestID = strings.TrimSpace(requestIDs[0])
	}
	if len(requestID) > 64 {
		return nil, bizerr.NewCode(CodeQueryFailed)
	}
	checkinMin, checkinMax, err := s.checkinRange(ctx)
	if err != nil {
		return nil, err
	}
	amount := grand.N(checkinMin, checkinMax)

	var result *CheckinResult
	err = dao.Checkin.Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {
		var player *entitymodel.User
		if txErr := dao.User.Ctx(ctx).Where(dao.User.Columns().Id, playerID).LockUpdate().Scan(&player); txErr != nil {
			return bizerr.WrapCode(txErr, CodeQueryFailed)
		}
		if player == nil {
			return bizerr.NewCode(CodeQueryFailed)
		}
		if requestID != "" {
			var replay *entitymodel.Checkin
			if txErr := dao.Checkin.Ctx(ctx).Where(do.Checkin{UserId: playerID, RequestId: requestID}).Scan(&replay); txErr != nil {
				return bizerr.WrapCode(txErr, CodeQueryFailed)
			}
			if replay != nil {
				result = &CheckinResult{Amount: replay.Amount, Balance: replay.ResultBalance}
				return nil
			}
		}
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
			RequestId:   requestID,
		}).InsertAndGetId()
		if txErr != nil {
			return bizerr.WrapCode(txErr, CodeWriteFailed)
		}

		newBalance, applyErr := s.ApplyDelta(ctx, tx, playerID, int64(amount), TxnTypeCheckin, checkinID)
		if applyErr != nil {
			return applyErr
		}
		if _, txErr = dao.Checkin.Ctx(ctx).Where(do.Checkin{Id: checkinID}).Data(do.Checkin{ResultBalance: newBalance}).Update(); txErr != nil {
			return bizerr.WrapCode(txErr, CodeWriteFailed)
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
func normalizeCheckinRange(minimum, maximum int) (int, int) {
	if minimum < 1 {
		minimum = 1
	}
	if maximum < minimum {
		maximum = minimum
	}
	return minimum, maximum
}
