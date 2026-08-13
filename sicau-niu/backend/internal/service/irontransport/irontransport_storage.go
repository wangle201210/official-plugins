// irontransport_storage.go contains transaction, locking and membership storage helpers.
package irontransport

import (
	"context"
	"encoding/json"
	"strings"
	"time"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/errors/gerror"

	"lina-core/pkg/bizerr"
	"lina-core/pkg/dialect"
	"lina-plugin-sicau-niu/backend/internal/dao"
	"lina-plugin-sicau-niu/backend/internal/model/do"
	entitymodel "lina-plugin-sicau-niu/backend/internal/model/entity"
)

func marshalStateSnapshot(state *State) (string, error) {
	data, err := json.Marshal(state)
	if err != nil {
		return "", bizerr.WrapCode(err, CodeWriteFailed)
	}
	return string(data), nil
}

func unmarshalStateSnapshot(value string) (*State, error) {
	if strings.TrimSpace(value) == "" {
		return nil, bizerr.NewCode(CodeQueryFailed)
	}
	var state State
	if err := json.Unmarshal([]byte(value), &state); err != nil {
		return nil, bizerr.WrapCode(err, CodeQueryFailed)
	}
	return &state, nil
}

func (s *serviceImpl) transaction(ctx context.Context, fn func(context.Context) error) error {
	return dao.TransportTeam.Transaction(ctx, func(ctx context.Context, _ gdb.TX) error { return fn(ctx) })
}

// lifecycleTransaction settles all bounded team expirations and runs one read
// projection under the same locks and authoritative timestamp. The callback
// returns only technical errors; callers defer business errors until commit so
// a rejected request cannot roll back lifecycle facts.
func (s *serviceImpl) lifecycleTransaction(ctx context.Context, fn func(context.Context, time.Time) error) error {
	return s.transaction(ctx, func(ctx context.Context) error {
		if err := lockTeamCapacity(ctx); err != nil {
			return bizerr.WrapCode(err, CodeQueryFailed)
		}
		_, now, err := s.expireInactiveLocked(ctx, time.Time{})
		if err != nil {
			return err
		}
		return fn(ctx, now)
	})
}

func lockPlayer(ctx context.Context, playerID int64) error {
	var row struct {
		ID int64 `json:"id"`
	}
	if err := dao.User.Ctx(ctx).Fields(dao.User.Columns().Id).Where(dao.User.Columns().Id, playerID).LockUpdate().Scan(&row); err != nil {
		return err
	}
	if row.ID <= 0 {
		return gerror.New("transport player not found")
	}
	return nil
}

func lockTeamCapacity(ctx context.Context) error {
	_, err := dao.TransportTeam.DB().Exec(ctx, "SELECT pg_advisory_xact_lock(?)", transportLockKey)
	return err
}

func effectiveTeamNameTaken(ctx context.Context, name string, excludeTeamID int64) (bool, error) {
	model := dao.TransportTeam.Ctx(ctx).Where(do.TransportTeam{Name: name, Status: teamStatusEffective})
	if excludeTeamID > 0 {
		model = model.WhereNot(dao.TransportTeam.Columns().Id, excludeTeamID)
	}
	count, err := model.Count()
	if err != nil {
		return false, bizerr.WrapCode(err, CodeQueryFailed)
	}
	return count > 0, nil
}

func transportWriteError(err error) error {
	if dialect.IsUniqueConstraintViolation(err) {
		return bizerr.NewCode(CodeTeamNameTaken)
	}
	return bizerr.WrapCode(err, CodeWriteFailed)
}

func activeMemberModel(ctx context.Context, playerID int64) *gdb.Model {
	return dao.TransportMember.Ctx(ctx).
		Where(dao.TransportMember.Columns().UserId, playerID).
		Where(dao.TransportMember.Columns().LeftAt + " IS NULL")
}

func activeMembership(ctx context.Context, playerID int64, lock bool) (*entitymodel.TransportMember, error) {
	model := activeMemberModel(ctx, playerID).Limit(1)
	if lock {
		model = model.LockUpdate()
	}
	var member *entitymodel.TransportMember
	if err := model.Scan(&member); err != nil {
		return nil, bizerr.WrapCode(err, CodeQueryFailed)
	}
	return member, nil
}

func (s *serviceImpl) invalidateLockedTeamIfExpired(ctx context.Context, team *entitymodel.TransportTeam, now time.Time) (bool, error) {
	if team == nil || team.LastActiveAt == nil || now.Before(team.LastActiveAt.Add(s.inactiveAfter)) {
		return false, nil
	}
	if _, err := dao.TransportTeam.Ctx(ctx).Where(do.TransportTeam{Id: team.Id}).Data(do.TransportTeam{
		Status: string(teamStatusInvalid), Visible: 0, MemberCount: 0,
		InvalidatedAt: &now, InvalidReason: string(invalidReasonIdle),
	}).Update(); err != nil {
		return false, bizerr.WrapCode(err, CodeWriteFailed)
	}
	if _, err := dao.TransportMember.Ctx(ctx).
		Where(dao.TransportMember.Columns().TeamId, team.Id).
		Where(dao.TransportMember.Columns().LeftAt + " IS NULL").
		Data(do.TransportMember{LeftAt: &now}).Update(); err != nil {
		return false, bizerr.WrapCode(err, CodeWriteFailed)
	}
	return true, nil
}
