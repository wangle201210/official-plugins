package grasssocial

import (
	"context"
	"sort"
	"time"

	"lina-core/pkg/bizerr"
	"lina-plugin-sicau-niu/backend/internal/dao"
	entitymodel "lina-plugin-sicau-niu/backend/internal/model/entity"
)

const activityCap = 20

// Activity is one structured social event that affected the current player.
type Activity struct {
	ID         int64
	ActorID    int64
	ActorName  string
	Amount     int
	OccurredAt *time.Time
	Type       string
}

// Activities returns recent events where the player was the steal target or
// gift recipient. Two bounded event reads and one user batch read keep query cost
// independent of the number of returned rows.
func (s *serviceImpl) Activities(ctx context.Context, playerID int64) ([]*Activity, error) {
	if playerID <= 0 {
		return []*Activity{}, nil
	}
	steals := make([]*entitymodel.Steal, 0)
	if err := dao.Steal.Ctx(ctx).
		Fields(dao.Steal.Columns().Id, dao.Steal.Columns().ActorUserId, dao.Steal.Columns().Amount, dao.Steal.Columns().CreatedAt).
		Where(dao.Steal.Columns().TargetUserId, playerID).
		OrderDesc(dao.Steal.Columns().Id).
		Limit(activityCap).
		Scan(&steals); err != nil {
		return nil, bizerr.WrapCode(err, CodeQueryFailed)
	}
	gifts := make([]*entitymodel.Gift, 0)
	if err := dao.Gift.Ctx(ctx).
		Fields(dao.Gift.Columns().Id, dao.Gift.Columns().FromUserId, dao.Gift.Columns().Amount, dao.Gift.Columns().CreatedAt).
		Where(dao.Gift.Columns().ToUserId, playerID).
		OrderDesc(dao.Gift.Columns().Id).
		Limit(activityCap).
		Scan(&gifts); err != nil {
		return nil, bizerr.WrapCode(err, CodeQueryFailed)
	}

	items := make([]*Activity, 0, len(steals)+len(gifts))
	actorIDs := make([]int64, 0, len(steals)+len(gifts))
	for _, row := range steals {
		items = append(items, &Activity{ID: row.Id, ActorID: row.ActorUserId, Amount: row.Amount, OccurredAt: row.CreatedAt, Type: "steal"})
		actorIDs = append(actorIDs, row.ActorUserId)
	}
	for _, row := range gifts {
		items = append(items, &Activity{ID: row.Id, ActorID: row.FromUserId, Amount: row.Amount, OccurredAt: row.CreatedAt, Type: "help"})
		actorIDs = append(actorIDs, row.FromUserId)
	}

	usersByID := make(map[int64]string, len(actorIDs))
	if len(actorIDs) > 0 {
		users := make([]*entitymodel.User, 0)
		if err := dao.User.Ctx(ctx).
			Fields(dao.User.Columns().Id, dao.User.Columns().Nickname).
			WhereIn(dao.User.Columns().Id, actorIDs).
			Scan(&users); err != nil {
			return nil, bizerr.WrapCode(err, CodeQueryFailed)
		}
		for _, user := range users {
			usersByID[user.Id] = user.Nickname
		}
	}
	for _, item := range items {
		item.ActorName = usersByID[item.ActorID]
	}
	sort.SliceStable(items, func(i, j int) bool {
		left, right := items[i].OccurredAt, items[j].OccurredAt
		if left == nil {
			return false
		}
		if right == nil {
			return true
		}
		return left.After(*right)
	})
	if len(items) > activityCap {
		items = items[:activityCap]
	}
	return items, nil
}
