// feeding_feed.go implements the feed action: it validates the target cattle is
// activated, computes the iron-cow proximity bonus from the cattle anchor and the
// current iron positions, and serializes writes for one player. The first
// successful response is persisted with the feeding so a client retry carrying
// the same request ID can replay it without applying the write twice.

package feeding

import (
	"context"
	"encoding/json"
	"time"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/util/grand"

	"lina-core/pkg/bizerr"
	"lina-plugin-sicau-niu/backend/internal/dao"
	"lina-plugin-sicau-niu/backend/internal/model/do"
	entitymodel "lina-plugin-sicau-niu/backend/internal/model/entity"
	"lina-plugin-sicau-niu/backend/internal/requestid"
	cattlesvc "lina-plugin-sicau-niu/backend/internal/service/cattle"
	"lina-plugin-sicau-niu/backend/internal/service/feeding/internal/ironlocation"
	grasssvc "lina-plugin-sicau-niu/backend/internal/service/grass"
)

// quoteEnabledOn is the enabled-flag value selecting quotes that participate in
// random playback, matching the catalog quote enabled flag.
const quoteEnabledOn = 1

// FeedInput defines the feed request.
type FeedInput struct {
	// NiuId is the target activated cattle ID to feed.
	NiuId int64
	// BaseAmount is the grass amount to feed (deducted from the balance).
	BaseAmount int
	// RequestId is the client idempotency key. A successful retry carrying the
	// same key replays the first result instead of deducting twice.
	RequestId string
}

// FeedOutput defines the result of a successful feeding.
type FeedOutput struct {
	// NiuId is the fed cattle ID.
	NiuId int64 `json:"niuId"`
	// NiuCode is the fed cattle serial code.
	NiuCode string `json:"niuCode"`
	// NiuName is the fed cattle name; empty for common cattle.
	NiuName string `json:"niuName"`
	// Quote is a random enabled school-history quote; empty when none exists.
	Quote string `json:"quote"`
	// BaseAmount is the original grass amount fed.
	BaseAmount int `json:"baseAmount"`
	// CoefficientBasis is the bonus coefficient in basis of 100 (100 or 150).
	CoefficientBasis int `json:"coefficientBasis"`
	// EffectAmount is the actual feeding effect = base * coefficient / 100.
	EffectAmount int `json:"effectAmount"`
	// IsIronBonus reports whether an iron-cow proximity bonus applied.
	IsIronBonus bool `json:"isIronBonus"`
	// Balance is the player's grass balance after the feeding deduction.
	Balance int64 `json:"balance"`
}

// Feed runs the validated, transactional feeding for playerID.
func (s *serviceImpl) Feed(ctx context.Context, playerID int64, in *FeedInput) (*FeedOutput, error) {
	if in == nil || in.NiuId <= 0 {
		return nil, bizerr.NewCode(CodeNiuIDRequired)
	}
	if in.BaseAmount <= 0 {
		return nil, bizerr.NewCode(CodeAmountInvalid)
	}
	requestID, requestOK := requestid.Normalize(in.RequestId)
	if !requestOK {
		return nil, bizerr.NewCode(CodeRequestIDRequired)
	}
	if playerID <= 0 {
		return nil, bizerr.NewCode(CodeQueryFailed)
	}
	replay, replayErr := s.feedByRequest(ctx, playerID, requestID)
	if replayErr != nil || replay != nil {
		return replay, replayErr
	}

	// Position reads may be backed by a replaceable gateway, so keep them outside
	// the player lock. Defer any error until after the in-transaction replay check:
	// a concurrent first request may already have committed the stable response.
	positions, positionsErr := s.ironLocation.Positions(ctx)
	var output *FeedOutput

	err := dao.Feeding.Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {
		if txErr := lockFeedPlayer(ctx, playerID); txErr != nil {
			return txErr
		}
		replay, txErr := s.feedByRequest(ctx, playerID, requestID)
		if txErr != nil {
			return txErr
		}
		if replay != nil {
			output = replay
			return nil
		}
		if positionsErr != nil {
			return positionsErr
		}
		niuRow, txErr := s.loadActiveNiu(ctx, in.NiuId)
		if txErr != nil {
			return txErr
		}
		isBonus, txErr := s.isIronBonusInRange(ctx, niuRow.Lat, niuRow.Lng, positions)
		if txErr != nil {
			return txErr
		}
		coefficientBasis := baseCoefficientBasis
		if isBonus {
			coefficientBasis = ironBonusCoefficientBasis
		}
		effectAmount := in.BaseAmount * coefficientBasis / 100
		isBonusFlag := 0
		if isBonus {
			isBonusFlag = 1
		}
		quote, txErr := s.randomEnabledQuote(ctx)
		if txErr != nil {
			return txErr
		}
		fedAt := time.Now()
		output = &FeedOutput{
			NiuId:            in.NiuId,
			NiuCode:          niuRow.Code,
			NiuName:          niuRow.Name,
			Quote:            quote,
			BaseAmount:       in.BaseAmount,
			CoefficientBasis: coefficientBasis,
			EffectAmount:     effectAmount,
			IsIronBonus:      isBonus,
		}
		feedingID, txErr := dao.Feeding.Ctx(ctx).Data(do.Feeding{
			UserId:           playerID,
			NiuId:            in.NiuId,
			BaseAmount:       in.BaseAmount,
			CoefficientBasis: coefficientBasis,
			EffectAmount:     effectAmount,
			IsIronBonus:      isBonusFlag,
			FedAt:            &fedAt,
			RequestId:        requestID,
		}).InsertAndGetId()
		if txErr != nil {
			return bizerr.WrapCode(txErr, CodeWriteFailed)
		}

		balance, applyErr := s.grassSvc.ApplyDelta(ctx, tx, playerID, -int64(in.BaseAmount), grasssvc.TxnTypeFeed, feedingID)
		if applyErr != nil {
			return applyErr
		}
		output.Balance = balance
		responseJSON, txErr := json.Marshal(output)
		if txErr != nil {
			return bizerr.WrapCode(txErr, CodeWriteFailed)
		}
		if _, txErr = dao.Feeding.Ctx(ctx).
			Where(do.Feeding{Id: feedingID}).
			Data(do.Feeding{ResponseJson: string(responseJSON)}).
			Update(); txErr != nil {
			return bizerr.WrapCode(txErr, CodeWriteFailed)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	return output, nil
}

// feedByRequest returns the response snapshot stored by the first successful
// request. A missing key returns nil so the caller can execute the write.
func (s *serviceImpl) feedByRequest(ctx context.Context, playerID int64, requestID string) (*FeedOutput, error) {
	var record *entitymodel.Feeding
	if err := dao.Feeding.Ctx(ctx).
		Fields(dao.Feeding.Columns().ResponseJson).
		Where(do.Feeding{UserId: playerID, RequestId: requestID}).
		Scan(&record); err != nil {
		return nil, bizerr.WrapCode(err, CodeQueryFailed)
	}
	if record == nil {
		return nil, nil
	}
	var output FeedOutput
	if err := json.Unmarshal([]byte(record.ResponseJson), &output); err != nil {
		return nil, bizerr.WrapCode(err, CodeQueryFailed)
	}
	return &output, nil
}

// lockFeedPlayer serializes all feed writes for one player, including distinct
// request IDs, and makes the request snapshot recheck race-free.
func lockFeedPlayer(ctx context.Context, playerID int64) error {
	var player *entitymodel.User
	if err := dao.User.Ctx(ctx).
		Fields(dao.User.Columns().Id).
		Where(do.User{Id: playerID}).
		LockUpdate().
		Scan(&player); err != nil {
		return bizerr.WrapCode(err, CodeQueryFailed)
	}
	if player == nil {
		return bizerr.NewCode(CodeQueryFailed)
	}
	return nil
}

// loadActiveNiu loads the target cattle and rejects it unless it is activated.
// It returns CodeNiuNotFound for a missing cattle and CodeNiuNotActive when the
// cattle has not been activated yet (C3 status).
func (s *serviceImpl) loadActiveNiu(ctx context.Context, niuID int64) (*entitymodel.Niu, error) {
	var niuRow *entitymodel.Niu
	err := dao.Niu.Ctx(ctx).
		Fields(
			dao.Niu.Columns().Code,
			dao.Niu.Columns().Name,
			dao.Niu.Columns().Lat,
			dao.Niu.Columns().Lng,
			dao.Niu.Columns().Status,
		).
		Where(do.Niu{Id: niuID}).
		Scan(&niuRow)
	if err != nil {
		return nil, bizerr.WrapCode(err, CodeQueryFailed)
	}
	if niuRow == nil {
		return nil, bizerr.NewCode(CodeNiuNotFound)
	}
	if niuRow.Status != cattlesvc.NiuStatusActive.String() {
		return nil, bizerr.NewCode(CodeNiuNotActive)
	}
	return niuRow, nil
}

// isIronBonusInRange reports whether the cattle anchor is within the configured
// proximity threshold of any prefetched iron-cow position. It tests the bounded
// position set in memory, so the check never issues a per-iron query.
func (s *serviceImpl) isIronBonusInRange(ctx context.Context, niuLat, niuLng float64, positions []*ironlocation.IronPosition) (bool, error) {
	threshold, err := s.currentIronBonusThreshold(ctx)
	if err != nil {
		return false, err
	}
	for _, position := range positions {
		if haversineMeters(niuLat, niuLng, position.Lat, position.Lng) < threshold {
			return true, nil
		}
	}
	return false, nil
}

// currentIronBonusThreshold returns the operator-maintained threshold when rules
// are injected, otherwise the constructor fallback.
func (s *serviceImpl) currentIronBonusThreshold(ctx context.Context) (float64, error) {
	if s.rulesSvc == nil {
		return s.ironBonusThreshold, nil
	}
	return s.rulesSvc.IronBonusThresholdMeters(ctx)
}

// randomEnabledQuote returns the content of one random enabled quote. It loads
// the enabled quote contents once and picks one uniformly with grand; an empty
// pool yields an empty string so feeding still succeeds without a quote.
func (s *serviceImpl) randomEnabledQuote(ctx context.Context) (string, error) {
	rows := make([]*entitymodel.Quote, 0)
	err := dao.Quote.Ctx(ctx).
		Fields(dao.Quote.Columns().Content).
		Where(do.Quote{Enabled: quoteEnabledOn}).
		Scan(&rows)
	if err != nil {
		return "", bizerr.WrapCode(err, CodeQueryFailed)
	}
	if len(rows) == 0 {
		return "", nil
	}
	return rows[grand.Intn(len(rows))].Content, nil
}
