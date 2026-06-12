// activation_guard.go implements the pre-transaction anti-cheat guards for
// photo check-ins: the daily attempt quota (counting failed attempts) that
// blocks virtual-location grid scanning, and the movement-speed guard that
// rejects implausible teleports between successive reported locations. Both
// guards read the operator-maintained runtime rules with built-in fallbacks,
// and neither failure response carries distance or bearing hints.

package activation

import (
	"context"
	"time"

	"lina-core/pkg/bizerr"
	"lina-plugin-sicau-niu/backend/internal/activityday"
	"lina-plugin-sicau-niu/backend/internal/dao"
	entitymodel "lina-plugin-sicau-niu/backend/internal/model/entity"
	rulessvc "lina-plugin-sicau-niu/backend/internal/service/rules"
)

// defaultDailyAttemptLimit and defaultMaxSpeedMps are the guard fallbacks used
// when no rules service is injected.
const (
	defaultDailyAttemptLimit = 20
	defaultMaxSpeedMps       = 25
)

// minSpeedWindow floors the elapsed time used by the speed computation so
// immediate retries with unchanged coordinates never divide by a near-zero
// interval while rapid scans across distant points still trip the guard.
const minSpeedWindow = time.Second

// activationGuards returns the operator-maintained anti-cheat guard values when
// the rules service is injected, otherwise the built-in fallbacks.
func (s *serviceImpl) activationGuards(ctx context.Context) (rulessvc.ActivationGuards, error) {
	if s.rulesSvc == nil {
		return rulessvc.ActivationGuards{
			DailyAttemptLimit: defaultDailyAttemptLimit,
			MaxSpeedMps:       defaultMaxSpeedMps,
		}, nil
	}
	return s.rulesSvc.ActivationGuards(ctx)
}

// guardDailyAttemptLimit rejects further photo check-ins once the player has
// used up the Beijing-day attempt quota; successes and failures both count.
// Quota rejections are not written to the attempt table, which keeps the audit
// volume bounded by the quota itself.
func (s *serviceImpl) guardDailyAttemptLimit(ctx context.Context, playerID int64, now time.Time, limit int) error {
	dayStart, dayEnd := activityday.DayBounds(now)
	columns := dao.ActivationAttempt.Columns()
	count, err := dao.ActivationAttempt.Ctx(ctx).
		Where(columns.UserId, playerID).
		WhereGTE(columns.AttemptedAt, dayStart).
		WhereLT(columns.AttemptedAt, dayEnd).
		Count()
	if err != nil {
		return bizerr.WrapCode(err, CodeQueryFailed)
	}
	if count >= limit {
		return bizerr.NewCode(CodeAttemptLimitReached)
	}
	return nil
}

// guardMovementSpeed rejects a check-in whose reported location implies
// implausible movement speed since the player's previous attempt
// (virtual-location teleport). The rejection writes a speed_anomaly attempt row
// as the risk record; the error itself carries no distance or bearing hints. A
// player with no prior attempt always passes.
func (s *serviceImpl) guardMovementSpeed(ctx context.Context, playerID int64, in *ActivateInput, now time.Time, maxSpeedMps int) error {
	columns := dao.ActivationAttempt.Columns()
	var lastRow *entitymodel.ActivationAttempt
	err := dao.ActivationAttempt.Ctx(ctx).
		Fields(columns.Lat, columns.Lng, columns.AttemptedAt).
		Where(columns.UserId, playerID).
		OrderDesc(columns.AttemptedAt).
		OrderDesc(columns.Id).
		Limit(1).
		Scan(&lastRow)
	if err != nil {
		return bizerr.WrapCode(err, CodeQueryFailed)
	}
	if lastRow == nil || lastRow.AttemptedAt == nil {
		return nil
	}

	elapsed := now.Sub(*lastRow.AttemptedAt)
	if elapsed < minSpeedWindow {
		elapsed = minSpeedWindow
	}
	movedMeters := haversineMeters(in.Lat, in.Lng, lastRow.Lat, lastRow.Lng)
	if movedMeters/elapsed.Seconds() <= float64(maxSpeedMps) {
		return nil
	}
	if err = insertActivationAttempt(ctx, playerID, in, 0, 0, activationAttemptSpeedAnomaly, 0, 0, now); err != nil {
		return err
	}
	return bizerr.NewCode(CodeSpeedAnomaly)
}
