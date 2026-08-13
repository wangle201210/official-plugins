// grasssocial_steal.go implements the steal action. It authorizes the target
// against the player's deterministic daily list and enforces the daily steal
// count limit (Beijing-time natural day). Writes for one player are serialized,
// and the first successful result is persisted so a retry carrying the same
// request ID does not apply the transfer or notification twice.

package grasssocial

import (
	"context"
	"fmt"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/util/grand"

	"lina-core/pkg/bizerr"
	"lina-plugin-sicau-niu/backend/internal/activityday"
	"lina-plugin-sicau-niu/backend/internal/dao"
	"lina-plugin-sicau-niu/backend/internal/model/do"
	entitymodel "lina-plugin-sicau-niu/backend/internal/model/entity"
	"lina-plugin-sicau-niu/backend/internal/requestid"
	grasssvc "lina-plugin-sicau-niu/backend/internal/service/grass"
)

// StealInput defines the steal request.
type StealInput struct {
	// TargetUserId is the target player ID to steal from.
	TargetUserId int64
	// RequestId is the client idempotency key. A successful retry carrying the
	// same key replays the first result instead of stealing twice.
	RequestId string
}

// StealResult is the outcome of a successful steal.
type StealResult struct {
	// Amount is the grass amount stolen and credited to the player.
	Amount int
	// Balance is the player's grass balance after the steal.
	Balance int64
}

// Steal runs the validated, transactional steal for playerID.
func (s *serviceImpl) Steal(ctx context.Context, playerID int64, in *StealInput) (*StealResult, error) {
	if in == nil || in.TargetUserId <= 0 {
		return nil, bizerr.NewCode(CodeTargetRequired)
	}
	if playerID <= 0 {
		return nil, bizerr.NewCode(CodeQueryFailed)
	}
	if in.TargetUserId == playerID {
		return nil, bizerr.NewCode(CodeSelfActionForbidden)
	}
	requestID, requestOK := requestid.Normalize(in.RequestId)
	if !requestOK {
		return nil, bizerr.NewCode(CodeRequestIDRequired)
	}
	if replay, replayErr := s.stealByRequest(ctx, playerID, requestID); replayErr != nil || replay != nil {
		return replay, replayErr
	}

	var (
		result *StealResult
		err    error
	)
	err = dao.Steal.Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {
		if txErr := lockSocialPlayers(ctx, playerID, in.TargetUserId); txErr != nil {
			return txErr
		}
		replay, txErr := s.stealByRequest(ctx, playerID, requestID)
		if txErr != nil {
			return txErr
		}
		if replay != nil {
			result = replay
			return nil
		}
		rules, txErr := s.socialRules(ctx)
		if txErr != nil {
			return txErr
		}
		today := activityday.Date(s.nowTime())
		authorized, txErr := s.targetInDailyList(ctx, playerID, in.TargetUserId, today, rules.StealDailyTargets)
		if txErr != nil {
			return txErr
		}
		if !authorized {
			return bizerr.NewCode(CodeTargetNotStealable)
		}
		if txErr = s.guardStealDailyLimit(ctx, playerID, today, rules.StealDailyLimit); txErr != nil {
			return txErr
		}
		targetBalance, txErr := s.currentBalance(ctx, in.TargetUserId)
		if txErr != nil {
			return txErr
		}
		if targetBalance <= 0 {
			return bizerr.NewCode(CodeTargetNoGrass)
		}

		amount := grand.N(rules.StealMinAmount, rules.StealMaxAmount)
		if int64(amount) > targetBalance {
			amount = int(targetBalance)
		}

		stealID, txErr := dao.Steal.Ctx(ctx).Data(do.Steal{
			ActorUserId:  playerID,
			TargetUserId: in.TargetUserId,
			Amount:       amount,
			StealDate:    today,
			RequestId:    requestID,
		}).InsertAndGetId()
		if txErr != nil {
			return bizerr.WrapCode(txErr, CodeWriteFailed)
		}

		if _, txErr = s.grassSvc.ApplyDelta(ctx, tx, in.TargetUserId, -int64(amount), grasssvc.TxnTypeStolenLoss, stealID); txErr != nil {
			return txErr
		}
		actorBalance, txErr := s.grassSvc.ApplyDelta(ctx, tx, playerID, int64(amount), grasssvc.TxnTypeStealGain, stealID)
		if txErr != nil {
			return txErr
		}
		if _, txErr = dao.Steal.Ctx(ctx).
			Where(do.Steal{Id: stealID}).
			Data(do.Steal{ResultBalance: actorBalance}).
			Update(); txErr != nil {
			return bizerr.WrapCode(txErr, CodeWriteFailed)
		}

		if txErr = s.insertInboxMessage(ctx, in.TargetUserId, MsgTypeStolen, stolenMessageContent(amount)); txErr != nil {
			return txErr
		}

		result = &StealResult{Amount: amount, Balance: actorBalance}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return result, nil
}

// stealByRequest returns the first successful result stored for requestID. A
// missing key returns nil so the caller can execute the write.
func (s *serviceImpl) stealByRequest(ctx context.Context, playerID int64, requestID string) (*StealResult, error) {
	var record *entitymodel.Steal
	if err := dao.Steal.Ctx(ctx).
		Fields(dao.Steal.Columns().Amount, dao.Steal.Columns().ResultBalance).
		Where(do.Steal{ActorUserId: playerID, RequestId: requestID}).
		Scan(&record); err != nil {
		return nil, bizerr.WrapCode(err, CodeQueryFailed)
	}
	if record == nil {
		return nil, nil
	}
	return &StealResult{Amount: record.Amount, Balance: record.ResultBalance}, nil
}

// targetInDailyList reports whether targetID appears in playerID's deterministic
// stealable list for the given day and authoritative rule snapshot.
func (s *serviceImpl) targetInDailyList(ctx context.Context, playerID, targetID int64, day string, dailyTargets int) (bool, error) {
	targets, err := s.dailyStealTargets(ctx, playerID, day, dailyTargets)
	if err != nil {
		return false, err
	}
	for _, target := range targets {
		if target.UserId == targetID {
			return true, nil
		}
	}
	return false, nil
}

// guardStealDailyLimit rejects a steal once the player reached dailyLimit. The
// caller supplies the authoritative rule snapshot read after player locking.
func (s *serviceImpl) guardStealDailyLimit(ctx context.Context, playerID int64, day string, dailyLimit int) error {
	count, err := dao.Steal.Ctx(ctx).
		Where(dao.Steal.Columns().ActorUserId, playerID).
		Where(dao.Steal.Columns().StealDate, day).
		Count()
	if err != nil {
		return bizerr.WrapCode(err, CodeQueryFailed)
	}
	if count >= dailyLimit {
		return bizerr.NewCode(CodeStealLimitReached)
	}
	return nil
}

// stolenMessageContent renders the stolen-notification body for the target inbox.
func stolenMessageContent(amount int) string {
	return fmt.Sprintf("你的草被偷走了 %d 份", amount)
}
