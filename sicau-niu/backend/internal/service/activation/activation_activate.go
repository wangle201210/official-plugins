// activation_activate.go implements the LBS activation action for mini-program
// photo check-ins. It validates the per-day success limit, the daily attempt
// quota and the movement-speed anti-cheat guard, matches the nearest currently
// visible inactive cattle within the configured distance threshold, then takes a
// per-cattle row lock (SELECT ... FOR UPDATE) inside the activation transaction
// before writing the first activation and flipping the cattle to active. The
// cattle main card is issued on success. Daily boundaries use the Beijing-time
// natural day from the activityday package.

package activation

import (
	"context"
	"sort"
	"time"

	"github.com/gogf/gf/v2/database/gdb"

	"lina-core/pkg/apitime"
	"lina-core/pkg/bizerr"
	"lina-plugin-sicau-niu/backend/internal/activityday"
	"lina-plugin-sicau-niu/backend/internal/dao"
	"lina-plugin-sicau-niu/backend/internal/model/do"
	entitymodel "lina-plugin-sicau-niu/backend/internal/model/entity"
	cattlesvc "lina-plugin-sicau-niu/backend/internal/service/cattle"
)

// firstActivatorFlag and laterActivatorFlag are the persisted is_first values.
const (
	firstActivatorFlag = 1
	laterActivatorFlag = 0
)

// activationAttemptResult is the stored audit result for a photo check-in.
type activationAttemptResult string

const (
	activationAttemptSuccess      activationAttemptResult = "success"
	activationAttemptNoNearby     activationAttemptResult = "no_nearby"
	activationAttemptOutOfRange   activationAttemptResult = "out_of_range"
	activationAttemptSpeedAnomaly activationAttemptResult = "speed_anomaly"
)

// activationCandidateCap bounds the server-side GPS matching query. The activity
// has at most 120 cattle, so this keeps the match path predictable without
// pagination.
const activationCandidateCap = 200

type activationMatch struct {
	niu          *entitymodel.Niu
	nearestNiuID int64
	distance     float64
	result       activationAttemptResult
}

// ActivateInput defines the LBS activation request.
type ActivateInput struct {
	// Lat is the player reported GPS latitude.
	Lat float64
	// Lng is the player reported GPS longitude.
	Lng float64
	// PhotoPath is the optional activation photo storage path (evidence only).
	PhotoPath string
}

// ActivateOutput defines the result of a successful activation.
type ActivateOutput struct {
	// NiuId is the server matched and activated cattle ID.
	NiuId int64
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
	if in == nil {
		return nil, bizerr.NewCode(CodeNoNearbyNiu)
	}
	if playerID <= 0 {
		return nil, bizerr.NewCode(CodeActivationNotFound)
	}

	if err := s.guardDailyLimit(ctx, playerID); err != nil {
		return nil, err
	}

	activatedAt := time.Now()
	activityDate := activityday.Date(activatedAt)

	guards, err := s.activationGuards(ctx)
	if err != nil {
		return nil, err
	}
	if err = s.guardDailyAttemptLimit(ctx, playerID, activatedAt, guards.DailyAttemptLimit); err != nil {
		return nil, err
	}
	if err = s.guardMovementSpeed(ctx, playerID, in, activatedAt, guards.MaxSpeedMps); err != nil {
		return nil, err
	}

	var output *ActivateOutput
	var activationErr error
	err = dao.Niu.Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {
		threshold, txErr := s.activationLBSThreshold(ctx)
		if txErr != nil {
			return txErr
		}
		match, txErr := matchAndLockNearbyInactiveNiu(ctx, in.Lat, in.Lng, activatedAt, threshold)
		if txErr != nil {
			return txErr
		}
		if match.niu == nil {
			if txErr = insertActivationAttempt(ctx, playerID, in, 0, match.nearestNiuID, match.result, match.distance, threshold, activatedAt); txErr != nil {
				return txErr
			}
			activationErr = bizerr.NewCode(CodeNoNearbyNiu)
			return nil
		}
		niuRow := match.niu
		niuID := niuRow.Id

		priorCount, txErr := dao.Activation.Ctx(ctx).
			Where(dao.Activation.Columns().NiuId, niuID).
			Count()
		if txErr != nil {
			return bizerr.WrapCode(txErr, CodeQueryFailed)
		}
		orderNo := priorCount + 1
		isFirst := priorCount == 0

		if isFirst {
			if _, txErr = dao.Niu.Ctx(ctx).
				Where(do.Niu{Id: niuID}).
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
			NiuId:        niuID,
			ActivityDate: activityDate,
			ActivatedAt:  &activatedAt,
			IsFirst:      isFirstFlag,
			OrderNo:      orderNo,
			PhotoPath:    in.PhotoPath,
		}).Insert(); txErr != nil {
			return bizerr.WrapCode(txErr, CodeWriteFailed)
		}
		if txErr = insertActivationAttempt(ctx, playerID, in, niuID, niuID, activationAttemptSuccess, match.distance, threshold, activatedAt); txErr != nil {
			return txErr
		}

		output = &ActivateOutput{
			NiuId:       niuID,
			IsFirst:     isFirst,
			OrderNo:     orderNo,
			ActivatedAt: apitime.MilliFromTime(activatedAt),
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	if activationErr != nil {
		return nil, activationErr
	}

	card, err := s.loadMainCard(ctx, output.NiuId)
	if err != nil {
		return nil, err
	}
	output.Card = card
	return output, nil
}

// activationLBSThreshold returns the operator-maintained LBS threshold when the
// rules service is injected, otherwise the constructor fallback.
func (s *serviceImpl) activationLBSThreshold(ctx context.Context) (float64, error) {
	if s.rulesSvc == nil {
		return s.lbsThreshold, nil
	}
	return s.rulesSvc.ActivationLBSThresholdMeters(ctx)
}

// guardDailyLimit rejects a second activation on the same Beijing-time natural
// day before entering the transaction. The active-set unique index back-stops
// this check against the concurrent race.
func (s *serviceImpl) guardDailyLimit(ctx context.Context, playerID int64) error {
	today := activityday.Today()
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
	return nil
}

// insertActivationAttempt records one photo check-in audit row. Daily-limit
// rejections intentionally call guardDailyLimit before this point and are not
// written to keep the audit table focused on location matching outcomes.
func insertActivationAttempt(
	ctx context.Context,
	playerID int64,
	in *ActivateInput,
	niuID int64,
	nearestNiuID int64,
	result activationAttemptResult,
	distance float64,
	threshold float64,
	attemptedAt time.Time,
) error {
	_, err := dao.ActivationAttempt.Ctx(ctx).Data(do.ActivationAttempt{
		UserId:       playerID,
		NiuId:        niuID,
		NearestNiuId: nearestNiuID,
		Result:       string(result),
		Lat:          in.Lat,
		Lng:          in.Lng,
		DistanceM:    distance,
		ThresholdM:   threshold,
		PhotoPath:    in.PhotoPath,
		AttemptedAt:  &attemptedAt,
	}).Insert()
	if err != nil {
		return bizerr.WrapCode(err, CodeWriteFailed)
	}
	return nil
}

// matchAndLockNearbyInactiveNiu finds the nearest currently visible inactive
// cattle within threshold and returns it locked. Concurrent activations may flip a
// candidate before this transaction locks it, so each locked row is rechecked and
// the matcher continues to the next candidate when that happens.
func matchAndLockNearbyInactiveNiu(
	ctx context.Context,
	lat, lng float64,
	now time.Time,
	threshold float64,
) (*activationMatch, error) {
	candidates, err := nearbyInactiveCandidates(ctx, now)
	if err != nil {
		return nil, err
	}

	type rankedCandidate struct {
		id       int64
		distance float64
	}
	ranked := make([]rankedCandidate, 0, len(candidates))
	for _, row := range candidates {
		if !niuCurrentlyVisible(row, now) {
			continue
		}
		distance := haversineMeters(lat, lng, row.Lat, row.Lng)
		ranked = append(ranked, rankedCandidate{id: row.Id, distance: distance})
	}
	if len(ranked) == 0 {
		return &activationMatch{result: activationAttemptNoNearby}, nil
	}

	sort.SliceStable(ranked, func(i, j int) bool {
		if ranked[i].distance == ranked[j].distance {
			return ranked[i].id < ranked[j].id
		}
		return ranked[i].distance < ranked[j].distance
	})
	nearest := ranked[0]
	if nearest.distance > threshold {
		return &activationMatch{
			nearestNiuID: nearest.id,
			distance:     nearest.distance,
			result:       activationAttemptOutOfRange,
		}, nil
	}

	for _, candidate := range ranked {
		if candidate.distance > threshold {
			break
		}
		niuRow, err := lockNiu(ctx, candidate.id)
		if err != nil {
			return nil, err
		}
		if niuRow.Status != cattlesvc.NiuStatusInactive.String() {
			continue
		}
		if !niuCurrentlyVisible(niuRow, now) {
			continue
		}
		distance := haversineMeters(lat, lng, niuRow.Lat, niuRow.Lng)
		if distance > threshold {
			continue
		}
		return &activationMatch{
			niu:          niuRow,
			nearestNiuID: candidate.id,
			distance:     distance,
			result:       activationAttemptSuccess,
		}, nil
	}
	return &activationMatch{
		nearestNiuID: nearest.id,
		distance:     nearest.distance,
		result:       activationAttemptNoNearby,
	}, nil
}

// nearbyInactiveCandidates returns a bounded, projected candidate set for GPS
// matching. Online time and inactive status are pushed to the database; optional
// weekday/time windows and exact Haversine distance are checked in memory.
func nearbyInactiveCandidates(ctx context.Context, now time.Time) ([]*entitymodel.Niu, error) {
	rows := make([]*entitymodel.Niu, 0)
	columns := dao.Niu.Columns()
	err := dao.Niu.Ctx(ctx).
		Fields(
			columns.Id,
			columns.Lat,
			columns.Lng,
			columns.OnlineAt,
			columns.VisibleWeekdays,
			columns.VisibleStart,
			columns.VisibleEnd,
			columns.Status,
		).
		Where(columns.OnlineAt+" IS NOT NULL").
		WhereLTE(columns.OnlineAt, now).
		Where(columns.Status, cattlesvc.NiuStatusInactive.String()).
		OrderAsc(columns.Id).
		Limit(activationCandidateCap).
		Scan(&rows)
	if err != nil {
		return nil, bizerr.WrapCode(err, CodeQueryFailed)
	}
	return rows, nil
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
