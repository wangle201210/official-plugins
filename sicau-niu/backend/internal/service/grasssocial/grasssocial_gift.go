// grasssocial_gift.go implements the gift action. It enforces the per-gift
// minimum amount, the recipient's existence and the daily gift count limit, then
// inside one transaction debits the giver (insufficient balance is rejected by
// the ledger), credits the recipient and writes a gift-received notification to
// the recipient's inbox.

package grasssocial

import (
	"context"
	"fmt"
	"time"

	"github.com/gogf/gf/v2/database/gdb"

	"lina-core/pkg/bizerr"
	"lina-plugin-sicau-niu/backend/internal/dao"
	"lina-plugin-sicau-niu/backend/internal/model/do"
	grasssvc "lina-plugin-sicau-niu/backend/internal/service/grass"
)

// GiftInput defines the gift request.
type GiftInput struct {
	// ToUserId is the recipient player ID.
	ToUserId int64
	// Amount is the grass amount to gift.
	Amount int
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
	if in.Amount < s.giftMinAmount {
		return nil, bizerr.NewCode(CodeGiftAmountTooSmall)
	}

	recipientExists, err := s.userExists(ctx, in.ToUserId)
	if err != nil {
		return nil, err
	}
	if !recipientExists {
		return nil, bizerr.NewCode(CodeRecipientNotFound)
	}

	today := time.Now().Format(socialDateLayout)
	if err = s.guardGiftDailyLimit(ctx, playerID, today); err != nil {
		return nil, err
	}

	var result *GiftResult
	err = dao.Gift.Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {
		giftID, txErr := dao.Gift.Ctx(ctx).Data(do.Gift{
			FromUserId: playerID,
			ToUserId:   in.ToUserId,
			Amount:     in.Amount,
			GiftDate:   today,
		}).InsertAndGetId()
		if txErr != nil {
			return bizerr.WrapCode(txErr, CodeWriteFailed)
		}

		giverBalance, txErr := s.grassSvc.ApplyDelta(ctx, tx, playerID, -int64(in.Amount), grasssvc.TxnTypeGiftOut, giftID)
		if txErr != nil {
			return txErr
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

// guardGiftDailyLimit rejects a gift once the player reached the daily count
// limit. The gift rows for the day back this count.
func (s *serviceImpl) guardGiftDailyLimit(ctx context.Context, playerID int64, day string) error {
	count, err := dao.Gift.Ctx(ctx).
		Where(dao.Gift.Columns().FromUserId, playerID).
		Where(dao.Gift.Columns().GiftDate, day).
		Count()
	if err != nil {
		return bizerr.WrapCode(err, CodeQueryFailed)
	}
	if count >= s.giftDailyLimit {
		return bizerr.NewCode(CodeGiftLimitReached)
	}
	return nil
}

// giftMessageContent renders the gift-received notification body for the
// recipient inbox.
func giftMessageContent(amount int) string {
	return fmt.Sprintf("你收到了 %d 份赠草", amount)
}
