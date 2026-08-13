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
	"encoding/json"
	"sort"
	"strings"
	"time"

	"github.com/gogf/gf/v2/database/gdb"

	"lina-core/pkg/apitime"
	"lina-core/pkg/bizerr"
	"lina-plugin-sicau-niu/backend/internal/activityday"
	"lina-plugin-sicau-niu/backend/internal/dao"
	"lina-plugin-sicau-niu/backend/internal/model/do"
	entitymodel "lina-plugin-sicau-niu/backend/internal/model/entity"
	"lina-plugin-sicau-niu/backend/internal/requestid"
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
	// RequestID is the mandatory player-scoped idempotency key supplied by the
	// mini program. Retrying it replays the first successful response.
	RequestID string
	// Lat is the player reported GPS latitude.
	Lat float64
	// Lng is the player reported GPS longitude.
	Lng float64
	// PhotoPath is the opaque player-owned activation photo identifier.
	PhotoPath string
}

// ActivateOutput defines the result of a successful activation.
type ActivateOutput struct {
	// Seq is the stable global activation sequence identifier.
	Seq int64 `json:"seq"`
	// NiuId is the server matched and activated cattle ID.
	NiuId int64 `json:"niuId"`
	// NiuName is the display name, falling back to the cattle code.
	NiuName string `json:"niuName"`
	// Skin is the deterministic mini-program cattle skin.
	Skin string `json:"skin"`
	// Quote is one enabled school-history quote selected at first execution.
	Quote string `json:"quote"`
	// IsFirst reports whether the player is the cattle first-activator.
	IsFirst bool `json:"isFirst"`
	// OrderNo is the player's arrival order for this cattle, starting at 1.
	OrderNo int `json:"orderNo"`
	// ActivatedAt is the activation time as a Unix timestamp in milliseconds.
	ActivatedAt *int64 `json:"activatedAt"`
	// Card is the cattle main card issued on activation; nil when none exists.
	Card *IssuedCard `json:"card"`
}

// IssuedCard defines the cattle main card returned on activation.
type IssuedCard struct {
	// Category is the card category string.
	Category string `json:"category"`
	// Title is the card title.
	Title string `json:"title"`
	// Content is the card content text.
	Content string `json:"content"`
	// ImagePath is the card image storage path; empty when none.
	ImagePath string `json:"imagePath"`
}

// Activate runs the validated, transactional LBS activation for playerID.
func (s *serviceImpl) Activate(ctx context.Context, playerID int64, in *ActivateInput) (*ActivateOutput, error) {
	requestID, err := activationRequestID(playerID, in)
	if err != nil {
		return nil, err
	}
	if replay, replayErr := s.activationByRequest(ctx, playerID, requestID); replayErr != nil || replay != nil {
		return replay, replayErr
	}
	var output *ActivateOutput
	var activationErr error
	err = dao.Niu.Transaction(ctx, func(ctx context.Context, _ gdb.TX) error {
		var player *entitymodel.User
		if txErr := dao.User.Ctx(ctx).Where(dao.User.Columns().Id, playerID).LockUpdate().Scan(&player); txErr != nil {
			return bizerr.WrapCode(txErr, CodeQueryFailed)
		}
		if player == nil {
			return bizerr.NewCode(CodeActivationNotFound)
		}
		replay, txErr := s.activationByRequest(ctx, playerID, requestID)
		if txErr != nil {
			return txErr
		}
		if replay != nil {
			output = replay
			return nil
		}
		// Validate photo ownership and unused state only after the same-player lock
		// and replay check. Concurrent retries of a committed request must replay
		// before observing that the first execution consumed its photo.
		if s.photoSvc != nil {
			if txErr = s.photoSvc.Validate(ctx, playerID, in.PhotoPath); txErr != nil {
				return txErr
			}
		}
		// Resolve every activation limit from one normalized rule-set read after
		// serialization and replay detection. A concurrent operator update can
		// therefore affect the next action, but cannot mix two rule versions in
		// this irreversible activation.
		rules, txErr := s.activationRuleSnapshot(ctx)
		if txErr != nil {
			return txErr
		}
		// Capture one authoritative time only after this player's writes are
		// serialized. Every daily guard, audit fact and activation row below uses
		// this same instant, including requests queued across Beijing midnight.
		activatedAt := s.nowTime()
		activityDate := activityday.Date(activatedAt)
		// The user row lock serializes distinct request IDs from the same player.
		// Recheck the daily guard here so a concurrent loser receives the stable
		// business error instead of leaking the backing unique-index error.
		if txErr := s.guardDailyLimit(ctx, playerID, activityDate); txErr != nil {
			return txErr
		}
		if txErr := s.guardDailyAttemptLimit(ctx, playerID, activatedAt, rules.dailyAttemptLimit); txErr != nil {
			return txErr
		}
		if txErr := s.guardMovementSpeed(ctx, playerID, in, activatedAt, rules.maxSpeedMps); txErr != nil {
			if bizerr.Is(txErr, CodeSpeedAnomaly) {
				// guardMovementSpeed wrote the audit row in this transaction. Commit
				// that risk record, then return the business rejection after commit.
				activationErr = txErr
				return nil
			}
			return txErr
		}
		threshold := rules.lbsThreshold
		match, txErr := matchAndLockNearbyNiu(ctx, playerID, in.Lat, in.Lng, activatedAt, threshold)
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
		activationID, txErr := dao.Activation.Ctx(ctx).Data(do.Activation{
			UserId:       playerID,
			NiuId:        niuID,
			ActivityDate: activityDate,
			ActivatedAt:  &activatedAt,
			IsFirst:      isFirstFlag,
			OrderNo:      orderNo,
			PhotoPath:    in.PhotoPath,
			RequestId:    requestID,
		}).InsertAndGetId()
		if txErr != nil {
			return bizerr.WrapCode(txErr, CodeWriteFailed)
		}
		if s.photoSvc != nil {
			if txErr = s.photoSvc.Consume(ctx, playerID, in.PhotoPath, activatedAt); txErr != nil {
				return txErr
			}
		}
		if txErr = insertActivationAttempt(ctx, playerID, in, niuID, niuID, activationAttemptSuccess, match.distance, threshold, activatedAt); txErr != nil {
			return txErr
		}

		output, txErr = s.buildActivationOutput(ctx, &entitymodel.Activation{
			Id: activationID, UserId: playerID, NiuId: niuID, IsFirst: isFirstFlag,
			OrderNo: orderNo, ActivatedAt: &activatedAt, RequestId: requestID,
		})
		if txErr != nil {
			return txErr
		}
		responseJSON, txErr := json.Marshal(output)
		if txErr != nil {
			return bizerr.WrapCode(txErr, CodeWriteFailed)
		}
		if _, txErr = dao.Activation.Ctx(ctx).Where(do.Activation{Id: activationID}).Data(do.Activation{
			ResponseJson: string(responseJSON),
		}).Update(); txErr != nil {
			return bizerr.WrapCode(txErr, CodeWriteFailed)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	if activationErr != nil {
		return nil, activationErr
	}

	return output, nil
}

func activationRequestID(playerID int64, in *ActivateInput) (string, error) {
	if in == nil {
		return "", bizerr.NewCode(CodeRequestIDRequired)
	}
	if playerID <= 0 {
		return "", bizerr.NewCode(CodeActivationNotFound)
	}
	requestID, ok := requestid.Normalize(in.RequestID)
	if !ok {
		return "", bizerr.NewCode(CodeRequestIDRequired)
	}
	return requestID, nil
}

func (s *serviceImpl) activationByRequest(ctx context.Context, playerID int64, requestID string) (*ActivateOutput, error) {
	var record *entitymodel.Activation
	if err := dao.Activation.Ctx(ctx).Unscoped().Where(do.Activation{UserId: playerID, RequestId: requestID}).Scan(&record); err != nil {
		return nil, bizerr.WrapCode(err, CodeQueryFailed)
	}
	if record == nil {
		return nil, nil
	}
	if strings.TrimSpace(record.ResponseJson) != "" {
		var out ActivateOutput
		if err := json.Unmarshal([]byte(record.ResponseJson), &out); err != nil {
			return nil, bizerr.WrapCode(err, CodeQueryFailed)
		}
		return &out, nil
	}
	return s.buildActivationOutput(ctx, record)
}

func (s *serviceImpl) buildActivationOutput(ctx context.Context, record *entitymodel.Activation) (*ActivateOutput, error) {
	var niu *entitymodel.Niu
	if err := dao.Niu.Ctx(ctx).Where(do.Niu{Id: record.NiuId}).Scan(&niu); err != nil {
		return nil, bizerr.WrapCode(err, CodeQueryFailed)
	}
	if niu == nil {
		return nil, bizerr.NewCode(CodeNiuNotFound)
	}
	quote, err := s.randomEnabledQuote(ctx)
	if err != nil {
		return nil, err
	}
	card, err := s.loadMainCard(ctx, record.NiuId)
	if err != nil {
		return nil, err
	}
	name := strings.TrimSpace(niu.Name)
	if name == "" {
		name = niu.Code
	}
	return &ActivateOutput{
		Seq: record.Id, NiuId: record.NiuId, NiuName: name, Skin: niuSkin(niu), Quote: quote,
		IsFirst: record.IsFirst == firstActivatorFlag, OrderNo: record.OrderNo,
		ActivatedAt: apitime.Milli(record.ActivatedAt), Card: card,
	}, nil
}

// guardDailyLimit rejects a second activation for the supplied authoritative
// Beijing business date. The caller holds the player row lock.
func (s *serviceImpl) guardDailyLimit(ctx context.Context, playerID int64, activityDate string) error {
	dailyCount, err := dao.Activation.Ctx(ctx).
		Where(dao.Activation.Columns().UserId, playerID).
		Where(dao.Activation.Columns().ActivityDate, activityDate).
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
// rejections are not written so the audit table stays focused on location
// matching outcomes.
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

// matchAndLockNearbyNiu finds the nearest visible cattle the player has not
// activated. Active cattle remain eligible for later visitors; the row lock
// serializes arrival-order assignment across concurrent players.
func matchAndLockNearbyNiu(
	ctx context.Context,
	playerID int64,
	lat, lng float64,
	now time.Time,
	threshold float64,
) (*activationMatch, error) {
	candidates, err := nearbyCandidates(ctx, playerID, now)
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
		if niuRow.Status != cattlesvc.NiuStatusInactive.String() && niuRow.Status != cattlesvc.NiuStatusActive.String() {
			continue
		}
		alreadyActivated, queryErr := dao.Activation.Ctx(ctx).
			Where(dao.Activation.Columns().UserId, playerID).
			Where(dao.Activation.Columns().NiuId, candidate.id).
			Count()
		if queryErr != nil {
			return nil, bizerr.WrapCode(queryErr, CodeQueryFailed)
		}
		if alreadyActivated > 0 {
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

// nearbyCandidates returns a bounded visible candidate set, excluding cattle the
// current player already activated in one batch query.
func nearbyCandidates(ctx context.Context, playerID int64, now time.Time) ([]*entitymodel.Niu, error) {
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
		WhereIn(columns.Status, []string{cattlesvc.NiuStatusInactive.String(), cattlesvc.NiuStatusActive.String()}).
		OrderAsc(columns.Id).
		Limit(activationCandidateCap).
		Scan(&rows)
	if err != nil {
		return nil, bizerr.WrapCode(err, CodeQueryFailed)
	}
	if playerID <= 0 || len(rows) == 0 {
		return rows, nil
	}
	ids := make([]int64, 0, len(rows))
	for _, row := range rows {
		ids = append(ids, row.Id)
	}
	activatedRows := make([]*entitymodel.Activation, 0)
	if err = dao.Activation.Ctx(ctx).
		Fields(dao.Activation.Columns().NiuId).
		Where(dao.Activation.Columns().UserId, playerID).
		WhereIn(dao.Activation.Columns().NiuId, ids).
		Scan(&activatedRows); err != nil {
		return nil, bizerr.WrapCode(err, CodeQueryFailed)
	}
	activated := make(map[int64]struct{}, len(activatedRows))
	for _, row := range activatedRows {
		activated[row.NiuId] = struct{}{}
	}
	filtered := make([]*entitymodel.Niu, 0, len(rows))
	for _, row := range rows {
		if _, exists := activated[row.Id]; !exists {
			filtered = append(filtered, row)
		}
	}
	return filtered, nil
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
