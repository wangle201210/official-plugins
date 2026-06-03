// grasssocial_steal.go implements the steal action. It authorizes the target
// against the player's deterministic daily list and enforces the daily steal
// count limit, then inside one transaction debits the target, credits the player
// and writes a stolen-notification to the target's inbox. The stolen amount is a
// small random value bounded by the target's current balance, so a steal never
// drives the target negative.

package grasssocial

import (
	"context"
	"fmt"
	"time"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/util/grand"

	"lina-core/pkg/bizerr"
	"lina-plugin-sicau-niu/backend/internal/dao"
	"lina-plugin-sicau-niu/backend/internal/model/do"
	grasssvc "lina-plugin-sicau-niu/backend/internal/service/grass"
)

// StealInput defines the steal request.
type StealInput struct {
	// TargetUserId is the target player ID to steal from.
	TargetUserId int64
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

	today := time.Now().Format(socialDateLayout)

	authorized, err := s.targetInDailyList(ctx, playerID, in.TargetUserId, today)
	if err != nil {
		return nil, err
	}
	if !authorized {
		return nil, bizerr.NewCode(CodeTargetNotStealable)
	}

	if err = s.guardStealDailyLimit(ctx, playerID, today); err != nil {
		return nil, err
	}

	var result *StealResult
	err = dao.Steal.Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {
		targetBalance, txErr := s.currentBalance(ctx, in.TargetUserId)
		if txErr != nil {
			return txErr
		}
		if targetBalance <= 0 {
			return bizerr.NewCode(CodeTargetNoGrass)
		}

		amount := grand.N(s.stealMinAmount, s.stealMaxAmount)
		if int64(amount) > targetBalance {
			amount = int(targetBalance)
		}

		stealID, txErr := dao.Steal.Ctx(ctx).Data(do.Steal{
			ActorUserId:  playerID,
			TargetUserId: in.TargetUserId,
			Amount:       amount,
			StealDate:    today,
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

// targetInDailyList reports whether targetID appears in playerID's deterministic
// stealable list for the given day, reusing the same derivation as StealTargets.
func (s *serviceImpl) targetInDailyList(ctx context.Context, playerID, targetID int64, day string) (bool, error) {
	targets, err := s.dailyStealTargets(ctx, playerID, day)
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

// guardStealDailyLimit rejects a steal once the player reached the daily count
// limit. The steal rows for the day back this count.
func (s *serviceImpl) guardStealDailyLimit(ctx context.Context, playerID int64, day string) error {
	count, err := dao.Steal.Ctx(ctx).
		Where(dao.Steal.Columns().ActorUserId, playerID).
		Where(dao.Steal.Columns().StealDate, day).
		Count()
	if err != nil {
		return bizerr.WrapCode(err, CodeQueryFailed)
	}
	if count >= s.stealDailyLimit {
		return bizerr.NewCode(CodeStealLimitReached)
	}
	return nil
}

// stolenMessageContent renders the stolen-notification body for the target inbox.
func stolenMessageContent(amount int) string {
	return fmt.Sprintf("你的草被偷走了 %d 份", amount)
}
