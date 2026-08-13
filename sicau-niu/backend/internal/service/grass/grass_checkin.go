// grass_checkin.go implements the daily check-in grant. A check-in inserts a
// uniquely-keyed (user_id, checkin_date) row keyed by the Beijing-time natural
// day, grants a random amount within the configured range and credits it to the
// player's ledger account, all inside one transaction so the check-in record,
// the credit and the balance stay consistent.

package grass

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/util/grand"

	"lina-core/pkg/bizerr"
	"lina-plugin-sicau-niu/backend/internal/activityday"
	"lina-plugin-sicau-niu/backend/internal/dao"
	"lina-plugin-sicau-niu/backend/internal/model/do"
	entitymodel "lina-plugin-sicau-niu/backend/internal/model/entity"
	"lina-plugin-sicau-niu/backend/internal/requestid"
)

// CheckinResult is the outcome of a successful daily check-in.
type CheckinResult struct {
	// Amount is the grass amount granted by this check-in.
	Amount int
	// Balance is the player's grass balance after the check-in credit.
	Balance int64
}

// Checkin grants the player a random daily check-in amount once per natural day.
// requestID is mandatory and replays the first successful response exactly.
func (s *serviceImpl) Checkin(ctx context.Context, playerID int64, requestID string) (*CheckinResult, error) {
	if playerID <= 0 {
		return nil, bizerr.NewCode(CodeQueryFailed)
	}

	var requestOK bool
	requestID, requestOK = requestid.Normalize(requestID)
	if !requestOK {
		return nil, bizerr.NewCode(CodeRequestIDRequired)
	}
	if replay, replayErr := s.checkinByRequest(ctx, playerID, requestID); replayErr != nil || replay != nil {
		return replay, replayErr
	}

	var result *CheckinResult
	err := dao.Checkin.Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {
		var player *entitymodel.User
		if txErr := dao.User.Ctx(ctx).Where(dao.User.Columns().Id, playerID).LockUpdate().Scan(&player); txErr != nil {
			return bizerr.WrapCode(txErr, CodeQueryFailed)
		}
		if player == nil {
			return bizerr.NewCode(CodeQueryFailed)
		}
		replay, txErr := s.checkinByRequest(ctx, playerID, requestID)
		if txErr != nil {
			return txErr
		}
		if replay != nil {
			result = replay
			return nil
		}
		// Derive the business day only after this player's writes are serialized,
		// so a request queued across Beijing midnight is checked and stored on one
		// authoritative day.
		today := activityday.Date(s.nowTime())
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
		checkinMin, checkinMax, txErr := s.checkinRange(ctx)
		if txErr != nil {
			return txErr
		}
		amount := grand.N(checkinMin, checkinMax)

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

// checkinByRequest returns the first successful result stored for requestID. A
// missing key returns nil so the caller can continue with a new write.
func (s *serviceImpl) checkinByRequest(ctx context.Context, playerID int64, requestID string) (*CheckinResult, error) {
	var record *entitymodel.Checkin
	if err := dao.Checkin.Ctx(ctx).
		Fields(dao.Checkin.Columns().Amount, dao.Checkin.Columns().ResultBalance).
		Where(do.Checkin{UserId: playerID, RequestId: requestID}).
		Scan(&record); err != nil {
		return nil, bizerr.WrapCode(err, CodeQueryFailed)
	}
	if record == nil {
		return nil, nil
	}
	return &CheckinResult{Amount: record.Amount, Balance: record.ResultBalance}, nil
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
