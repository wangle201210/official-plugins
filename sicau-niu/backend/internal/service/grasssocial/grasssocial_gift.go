// grasssocial_gift.go implements the gift action. It enforces the per-gift
// minimum amount, the recipient's existence and the daily gift count limit
// (Beijing-time natural day). Writes for one giver are serialized, and the first
// successful result is persisted so a retry carrying the same request ID does
// not apply the transfer or notification twice.

package grasssocial

import (
	"context"
	"fmt"

	"github.com/gogf/gf/v2/database/gdb"

	"lina-core/pkg/bizerr"
	"lina-plugin-sicau-niu/backend/internal/activityday"
	"lina-plugin-sicau-niu/backend/internal/dao"
	"lina-plugin-sicau-niu/backend/internal/model/do"
	entitymodel "lina-plugin-sicau-niu/backend/internal/model/entity"
	"lina-plugin-sicau-niu/backend/internal/requestid"
	grasssvc "lina-plugin-sicau-niu/backend/internal/service/grass"
)

// GiftInput defines the gift request.
type GiftInput struct {
	// ToUserId is the recipient player ID.
	ToUserId int64
	// Amount is the grass amount to gift.
	Amount int
	// RequestId is the client idempotency key. A successful retry carrying the
	// same key replays the first result instead of gifting twice.
	RequestId string
}

// GiftResult is the outcome of a successful gift.
type GiftResult struct {
	// Balance is the giver's grass balance after the gift.
	Balance int64
}

// Gift runs the validated, transactional gift for playerID.
func (s *serviceImpl) Gift(ctx context.Context, playerID int64, in *GiftInput) (*GiftResult, error) {
	if in == nil || in.ToUserId <= 0 {
		return nil, bizerr.NewCode(CodeTargetRequired)
	}
	if playerID <= 0 {
		return nil, bizerr.NewCode(CodeQueryFailed)
	}
	if in.ToUserId == playerID {
		return nil, bizerr.NewCode(CodeSelfActionForbidden)
	}
	requestID, requestOK := requestid.Normalize(in.RequestId)
	if !requestOK {
		return nil, bizerr.NewCode(CodeRequestIDRequired)
	}
	if replay, replayErr := s.giftByRequest(ctx, playerID, requestID); replayErr != nil || replay != nil {
		return replay, replayErr
	}
	recipientExists, err := s.userExists(ctx, in.ToUserId)
	if err != nil {
		return nil, err
	}
	if !recipientExists {
		return nil, bizerr.NewCode(CodeRecipientNotFound)
	}

	var result *GiftResult
	err = dao.Gift.Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {
		if txErr := lockSocialPlayers(ctx, playerID, in.ToUserId); txErr != nil {
			return txErr
		}
		replay, txErr := s.giftByRequest(ctx, playerID, requestID)
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
		if in.Amount < rules.GiftMinAmount {
			return bizerr.NewCode(CodeGiftAmountTooSmall)
		}
		today := activityday.Date(s.nowTime())
		if txErr = s.guardGiftDailyLimit(ctx, playerID, today, rules.GiftDailyLimit); txErr != nil {
			return txErr
		}
		giftID, txErr := dao.Gift.Ctx(ctx).Data(do.Gift{
			FromUserId: playerID,
			ToUserId:   in.ToUserId,
			Amount:     in.Amount,
			GiftDate:   today,
			RequestId:  requestID,
		}).InsertAndGetId()
		if txErr != nil {
			return bizerr.WrapCode(txErr, CodeWriteFailed)
		}

		giverBalance, txErr := s.grassSvc.ApplyDelta(ctx, tx, playerID, -int64(in.Amount), grasssvc.TxnTypeGiftOut, giftID)
		if txErr != nil {
			return txErr
		}
		if _, txErr = dao.Gift.Ctx(ctx).
			Where(do.Gift{Id: giftID}).
			Data(do.Gift{ResultBalance: giverBalance}).
			Update(); txErr != nil {
			return bizerr.WrapCode(txErr, CodeWriteFailed)
		}
		if _, txErr = s.grassSvc.ApplyDelta(ctx, tx, in.ToUserId, int64(in.Amount), grasssvc.TxnTypeGiftIn, giftID); txErr != nil {
			return txErr
		}

		if txErr = s.insertInboxMessage(ctx, in.ToUserId, MsgTypeGiftReceived, giftMessageContent(in.Amount)); txErr != nil {
			return txErr
		}

		result = &GiftResult{Balance: giverBalance}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return result, nil
}

// giftByRequest returns the first successful result stored for requestID. A
// missing key returns nil so the caller can execute the write.
func (s *serviceImpl) giftByRequest(ctx context.Context, playerID int64, requestID string) (*GiftResult, error) {
	var record *entitymodel.Gift
	if err := dao.Gift.Ctx(ctx).
		Fields(dao.Gift.Columns().ResultBalance).
		Where(do.Gift{FromUserId: playerID, RequestId: requestID}).
		Scan(&record); err != nil {
		return nil, bizerr.WrapCode(err, CodeQueryFailed)
	}
	if record == nil {
		return nil, nil
	}
	return &GiftResult{Balance: record.ResultBalance}, nil
}

// guardGiftDailyLimit rejects a gift once the player reached dailyLimit. The
// caller supplies the authoritative rule snapshot read after player locking.
func (s *serviceImpl) guardGiftDailyLimit(ctx context.Context, playerID int64, day string, dailyLimit int) error {
	count, err := dao.Gift.Ctx(ctx).
		Where(dao.Gift.Columns().FromUserId, playerID).
		Where(dao.Gift.Columns().GiftDate, day).
		Count()
	if err != nil {
		return bizerr.WrapCode(err, CodeQueryFailed)
	}
	if count >= dailyLimit {
		return bizerr.NewCode(CodeGiftLimitReached)
	}
	return nil
}

// giftMessageContent renders the gift-received notification body for the
// recipient inbox.
func giftMessageContent(amount int) string {
	return fmt.Sprintf("你收到了 %d 份赠草", amount)
}
