// activation_activate.go implements the LBS activation action with shared-pool
// first-activator concurrency. It validates the reported location against the
// configured LBS threshold, the per-day limit and the no-duplicate rule, then
// inside a transaction takes a per-cattle row lock (SELECT ... FOR UPDATE) to
// serialize concurrent activations of the same cattle so exactly one player
// becomes the first activator. The first activator flips the cattle to active and
// arrival order starts at 1; later activators get an incremented order. The
// cattle main card is issued on success.

package activation

import (
	"context"
	"time"

	"github.com/gogf/gf/v2/database/gdb"

	"lina-core/pkg/apitime"
	"lina-core/pkg/bizerr"
	"lina-plugin-sicau-niu/backend/internal/dao"
	"lina-plugin-sicau-niu/backend/internal/model/do"
	entitymodel "lina-plugin-sicau-niu/backend/internal/model/entity"
	cattlesvc "lina-plugin-sicau-niu/backend/internal/service/cattle"
)

// activityDateLayout is the YYYY-MM-DD natural-day key used for the per-day limit.
const activityDateLayout = "2006-01-02"

// firstActivatorFlag and laterActivatorFlag are the persisted is_first values.
const (
	firstActivatorFlag = 1
	laterActivatorFlag = 0
)

// ActivateInput defines the LBS activation request.
type ActivateInput struct {
	// NiuId is the target cattle ID to activate.
	NiuId int64
	// Lat is the player reported GPS latitude.
	Lat float64
	// Lng is the player reported GPS longitude.
	Lng float64
	// PhotoPath is the optional activation photo storage path (evidence only).
	PhotoPath string
}

// ActivateOutput defines the result of a successful activation.
type ActivateOutput struct {
	// IsFirst reports whether the player is the cattle first-activator.
	IsFirst bool
	// OrderNo is the player's arrival order for this cattle, starting at 1.
	OrderNo int
	// ActivatedAt is the activation time as a Unix timestamp in milliseconds.
	ActivatedAt *int64
	// Card is the cattle main card issued on activation; nil when none exists.
	Card *IssuedCard
}

// IssuedCard defines the cattle main card returned on activation.
type IssuedCard struct {
	// Category is the card category string.
	Category string
	// Title is the card title.
	Title string
	// Content is the card content text.
	Content string
	// ImagePath is the card image storage path; empty when none.
	ImagePath string
}

// Activate runs the validated, transactional LBS activation for playerID.
func (s *serviceImpl) Activate(ctx context.Context, playerID int64, in *ActivateInput) (*ActivateOutput, error) {
	if in == nil || in.NiuId <= 0 {
		return nil, bizerr.NewCode(CodeNiuIDRequired)
	}
	if playerID <= 0 {
		return nil, bizerr.NewCode(CodeActivationNotFound)
	}

	if err := s.guardDailyAndDuplicate(ctx, playerID, in.NiuId); err != nil {
		return nil, err
	}

	activatedAt := time.Now()
	activityDate := activatedAt.Format(activityDateLayout)

	var output *ActivateOutput
	err := dao.Niu.Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {
		niuRow, txErr := lockNiu(ctx, in.NiuId)
		if txErr != nil {
			return txErr
		}
		if distance := haversineMeters(in.Lat, in.Lng, niuRow.Lat, niuRow.Lng); distance > s.lbsThreshold {
			return bizerr.NewCode(CodeOutOfRange)
		}

		priorCount, txErr := dao.Activation.Ctx(ctx).
			Where(dao.Activation.Columns().NiuId, in.NiuId).
			Count()
		if txErr != nil {
			return bizerr.WrapCode(txErr, CodeQueryFailed)
		}
		orderNo := priorCount + 1
		isFirst := priorCount == 0

		if isFirst {
			if _, txErr = dao.Niu.Ctx(ctx).
				Where(do.Niu{Id: in.NiuId}).
				Data(do.Niu{Status: cattlesvc.NiuStatusActive.String()}).
				Update(); txErr != nil {
				return bizerr.WrapCode(txErr, CodeWriteFailed)
			}
		}

		isFirstFlag := laterActivatorFlag
		if isFirst {
			isFirstFlag = firstActivatorFlag
		}
		if _, txErr = dao.Activation.Ctx(ctx).Data(do.Activation{
			UserId:       playerID,
			NiuId:        in.NiuId,
			ActivityDate: activityDate,
			ActivatedAt:  &activatedAt,
			IsFirst:      isFirstFlag,
			OrderNo:      orderNo,
			PhotoPath:    in.PhotoPath,
		}).Insert(); txErr != nil {
			return bizerr.WrapCode(txErr, CodeWriteFailed)
		}

		output = &ActivateOutput{
			IsFirst:     isFirst,
			OrderNo:     orderNo,
			ActivatedAt: apitime.MilliFromTime(activatedAt),
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	card, err := s.loadMainCard(ctx, in.NiuId)
	if err != nil {
		return nil, err
	}
	output.Card = card
	return output, nil
}

// guardDailyAndDuplicate rejects a second activation on the same natural day and a
// repeat activation of the same cattle before entering the transaction. The
// active-set unique indexes back-stop these checks against the concurrent race.
func (s *serviceImpl) guardDailyAndDuplicate(ctx context.Context, playerID, niuID int64) error {
	today := time.Now().Format(activityDateLayout)
	dailyCount, err := dao.Activation.Ctx(ctx).
		Where(dao.Activation.Columns().UserId, playerID).
		Where(dao.Activation.Columns().ActivityDate, today).
		Count()
	if err != nil {
		return bizerr.WrapCode(err, CodeQueryFailed)
	}
	if dailyCount > 0 {
		return bizerr.NewCode(CodeDailyLimitReached)
	}

	duplicateCount, err := dao.Activation.Ctx(ctx).
		Where(dao.Activation.Columns().UserId, playerID).
		Where(dao.Activation.Columns().NiuId, niuID).
		Count()
	if err != nil {
		return bizerr.WrapCode(err, CodeQueryFailed)
	}
	if duplicateCount > 0 {
		return bizerr.NewCode(CodeAlreadyActivated)
	}
	return nil
}

// lockNiu loads the target cattle row under a row lock (SELECT ... FOR UPDATE)
// inside the activation transaction so concurrent activations of the same cattle
// are serialized. It returns CodeNiuNotFound when the cattle is missing.
func lockNiu(ctx context.Context, niuID int64) (*entitymodel.Niu, error) {
	var niuRow *entitymodel.Niu
	err := dao.Niu.Ctx(ctx).
		Where(do.Niu{Id: niuID}).
		LockUpdate().
		Scan(&niuRow)
	if err != nil {
		return nil, bizerr.WrapCode(err, CodeQueryFailed)
	}
	if niuRow == nil {
		return nil, bizerr.NewCode(CodeNiuNotFound)
	}
	return niuRow, nil
}

// loadMainCard loads the cattle's unique active main card for issuance. A cattle
// without a main card yields a nil card so activation still succeeds.
func (s *serviceImpl) loadMainCard(ctx context.Context, niuID int64) (*IssuedCard, error) {
	var cardRow *entitymodel.Card
	err := dao.Card.Ctx(ctx).Where(do.Card{NiuId: niuID}).Scan(&cardRow)
	if err != nil {
		return nil, bizerr.WrapCode(err, CodeQueryFailed)
	}
	if cardRow == nil {
		return nil, nil
	}
	return &IssuedCard{
		Category:  cardRow.Category,
		Title:     cardRow.Title,
		Content:   cardRow.Content,
		ImagePath: cardRow.ImagePath,
	}, nil
}
