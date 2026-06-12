// grasssocial_steal_targets.go implements the deterministic daily stealable
// list. For a given (player, natural day) it derives a stable seed, loads the
// other players once, deterministically shuffles them with that seed and takes
// the configured number of targets. No candidate list is stored: the same list
// can be recomputed to authorize a steal, so steal validation reuses this exact
// derivation.

package grasssocial

import (
	"context"
	"math/rand"

	"lina-core/pkg/bizerr"
	"lina-plugin-sicau-niu/backend/internal/activityday"
	"lina-plugin-sicau-niu/backend/internal/dao"
	entitymodel "lina-plugin-sicau-niu/backend/internal/model/entity"
)

// candidateScanLimit bounds the other-player set loaded for the daily shuffle so
// the list derivation reads a bounded, stable candidate pool rather than the
// entire (potentially large) player table.
const candidateScanLimit = 500

// StealTarget is one stealable player projected for the daily list.
type StealTarget struct {
	// UserId is the stealable target player ID.
	UserId int64
	// Nickname is the target player nickname.
	Nickname string
}

// StealTargets returns the player's deterministic daily stealable list.
func (s *serviceImpl) StealTargets(ctx context.Context, playerID int64) ([]*StealTarget, error) {
	if playerID <= 0 {
		return nil, bizerr.NewCode(CodeQueryFailed)
	}
	return s.dailyStealTargets(ctx, playerID, activityday.Today())
}

// dailyStealTargets computes the stealable list for a player on a specific day.
// It is shared by StealTargets and the steal authorization so both observe the
// same deterministic membership.
func (s *serviceImpl) dailyStealTargets(ctx context.Context, playerID int64, day string) ([]*StealTarget, error) {
	rows := make([]*entitymodel.User, 0)
	err := dao.User.Ctx(ctx).
		Fields(
			dao.User.Columns().Id,
			dao.User.Columns().Nickname,
		).
		WhereNot(dao.User.Columns().Id, playerID).
		OrderAsc(dao.User.Columns().Id).
		Limit(candidateScanLimit).
		Scan(&rows)
	if err != nil {
		return nil, bizerr.WrapCode(err, CodeQueryFailed)
	}
	if len(rows) == 0 {
		return []*StealTarget{}, nil
	}

	order := deterministicOrder(len(rows), dailySeed(playerID, day))
	rules, err := s.socialRules(ctx)
	if err != nil {
		return nil, err
	}
	count := rules.StealDailyTargets
	if count > len(rows) {
		count = len(rows)
	}

	targets := make([]*StealTarget, 0, count)
	for i := 0; i < count; i++ {
		row := rows[order[i]]
		targets = append(targets, &StealTarget{UserId: row.Id, Nickname: row.Nickname})
	}
	return targets, nil
}

// dailySeed derives a stable 64-bit seed from the player ID and the natural-day
// key so the stealable list is reproducible for the whole day and differs per
// player and per day.
func dailySeed(playerID int64, day string) int64 {
	// FNV-1a 64-bit mix computed inline (no io.Writer error returns to handle),
	// so the stealable list is reproducible per player per day.
	const (
		offset64 = uint64(14695981039346656037)
		prime64  = uint64(1099511628211)
	)
	hash := offset64
	for i := 0; i < len(day); i++ {
		hash = (hash ^ uint64(day[i])) * prime64
	}
	value := uint64(playerID)
	for shift := uint(0); shift < 64; shift += 8 {
		hash = (hash ^ ((value >> shift) & 0xff)) * prime64
	}
	return int64(hash)
}

// deterministicOrder returns a permutation of [0,n) produced by a seeded PRNG, so
// the same seed always yields the same shuffle. A seeded *rand.Rand is used
// deliberately for reproducibility; it is not the global argless RNG.
func deterministicOrder(n int, seed int64) []int {
	return rand.New(rand.NewSource(seed)).Perm(n)
}
