// feeding_feed.go implements the feed action: it validates the target cattle is
// activated, computes the iron-cow proximity bonus from the cattle anchor and the
// current iron positions, deduplicates client retries by the optional request ID,
// then inside one transaction debits the player's ledger by the base amount and
// records the feeding with the original amount, the bonus coefficient and the
// resulting effect. The response carries the cattle info, a random enabled
// school-history quote and the bonus breakdown.

package feeding

import (
	"context"
	"time"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/util/grand"

	"lina-core/pkg/bizerr"
	"lina-plugin-sicau-niu/backend/internal/dao"
	"lina-plugin-sicau-niu/backend/internal/model/do"
	entitymodel "lina-plugin-sicau-niu/backend/internal/model/entity"
	cattlesvc "lina-plugin-sicau-niu/backend/internal/service/cattle"
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
	// RequestId is the optional client idempotency key; a retry carrying an
	// already-recorded key is rejected as a duplicate instead of deducting twice.
	RequestId string
}

// FeedOutput defines the result of a successful feeding.
type FeedOutput struct {
	// NiuId is the fed cattle ID.
	NiuId int64
	// NiuCode is the fed cattle serial code.
	NiuCode string
	// NiuName is the fed cattle name; empty for common cattle.
	NiuName string
	// Quote is a random enabled school-history quote; empty when none exists.
	Quote string
	// BaseAmount is the original grass amount fed.
	BaseAmount int
	// CoefficientBasis is the bonus coefficient in basis of 100 (100 or 150).
	CoefficientBasis int
	// EffectAmount is the actual feeding effect = base * coefficient / 100.
	EffectAmount int
	// IsIronBonus reports whether an iron-cow proximity bonus applied.
	IsIronBonus bool
	// Balance is the player's grass balance after the feeding deduction.
	Balance int64
}

// Feed runs the validated, transactional feeding for playerID.
func (s *serviceImpl) Feed(ctx context.Context, playerID int64, in *FeedInput) (*FeedOutput, error) {
	if in == nil || in.NiuId <= 0 {
		return nil, bizerr.NewCode(CodeNiuIDRequired)
	}
	if in.BaseAmount <= 0 {
		return nil, bizerr.NewCode(CodeAmountInvalid)
	}
	if playerID <= 0 {
		return nil, bizerr.NewCode(CodeQueryFailed)
	}

	niuRow, err := s.loadActiveNiu(ctx, in.NiuId)
	if err != nil {
		return nil, err
	}

	isBonus, err := s.isIronBonusInRange(ctx, niuRow.Lat, niuRow.Lng)
	if err != nil {
		return nil, err
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
	fedAt := time.Now()

	var newBalance int64
	err = dao.Feeding.Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {
		if txErr := s.guardDuplicateFeed(ctx, playerID, in.RequestId); txErr != nil {
			return txErr
		}
		feedingID, txErr := dao.Feeding.Ctx(ctx).Data(do.Feeding{
			UserId:           playerID,
			NiuId:            in.NiuId,
			BaseAmount:       in.BaseAmount,
			CoefficientBasis: coefficientBasis,
			EffectAmount:     effectAmount,
			IsIronBonus:      isBonusFlag,
			FedAt:            &fedAt,
			RequestId:        in.RequestId,
		}).InsertAndGetId()
		if txErr != nil {
			return bizerr.WrapCode(txErr, CodeWriteFailed)
		}

		balance, applyErr := s.grassSvc.ApplyDelta(ctx, tx, playerID, -int64(in.BaseAmount), grasssvc.TxnTypeFeed, feedingID)
		if applyErr != nil {
			return applyErr
		}
		newBalance = balance
		return nil
	})
	if err != nil {
		return nil, err
	}

	quote, err := s.randomEnabledQuote(ctx)
	if err != nil {
		return nil, err
	}

	return &FeedOutput{
		NiuId:            in.NiuId,
		NiuCode:          niuRow.Code,
		NiuName:          niuRow.Name,
		Quote:            quote,
		BaseAmount:       in.BaseAmount,
		CoefficientBasis: coefficientBasis,
		EffectAmount:     effectAmount,
		IsIronBonus:      isBonus,
		Balance:          newBalance,
	}, nil
}

// guardDuplicateFeed rejects a feed whose non-empty request ID was already
// recorded for the player. It runs inside the feeding transaction; the partial
// unique index on (user_id, request_id) back-stops the concurrent race.
func (s *serviceImpl) guardDuplicateFeed(ctx context.Context, playerID int64, requestID string) error {
	if requestID == "" {
		return nil
	}
	count, err := dao.Feeding.Ctx(ctx).
		Where(dao.Feeding.Columns().UserId, playerID).
		Where(dao.Feeding.Columns().RequestId, requestID).
		Count()
	if err != nil {
		return bizerr.WrapCode(err, CodeQueryFailed)
	}
	if count > 0 {
		return bizerr.NewCode(CodeDuplicateRequest)
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
// proximity threshold of any iron cow's current position. It loads all current
// iron positions once and tests each in memory, so the check never issues a
// per-iron query.
func (s *serviceImpl) isIronBonusInRange(ctx context.Context, niuLat, niuLng float64) (bool, error) {
	positions, err := s.ironLocation.Positions(ctx)
	if err != nil {
		return false, err
	}
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
