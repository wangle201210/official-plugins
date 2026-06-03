// honor_player.go implements the read-only player honor unlock computation. It
// pre-aggregates, in a fixed number of batched queries, every count the unlock
// rules need: the player's feeding count, the player's activation count, the
// player's collected-card counts per category (the main cards of the cattle the
// player activated, grouped by card category) and the active-card totals per
// category. Each honor definition is then evaluated in memory against those
// pre-aggregated counts, so the cost is independent of the number of honor
// definitions and never issues a per-honor query. The computation is read-only:
// it never writes the user_honor grant table, which the C7 settlement owns.

package honor

import (
	"context"

	"lina-core/pkg/bizerr"
	"lina-plugin-sicau-niu/backend/internal/dao"
	entitymodel "lina-plugin-sicau-niu/backend/internal/model/entity"
)

// PlayerHonorItem defines one honor definition with the requesting player's
// read-only unlock status.
type PlayerHonorItem struct {
	Id         int64
	HonorType  string
	Code       string
	Name       string
	UnlockType string
	Threshold  int
	Category   string
	ImagePath  string
	Unlocked   bool
}

// playerProgress holds the player's pre-aggregated counts used to evaluate every
// unlock rule in memory.
type playerProgress struct {
	// feedCount is the number of feeding records the player has.
	feedCount int
	// activationCount is the number of activation records the player has.
	activationCount int
	// collectedByCategory maps a card category to the number of active cards in
	// that category the player has collected.
	collectedByCategory map[string]int
	// collectedTotal is the total number of active cards the player has collected.
	collectedTotal int
	// activeByCategory maps a card category to the number of active cards in that
	// category overall.
	activeByCategory map[string]int
	// activeTotal is the total number of active cards overall.
	activeTotal int
}

// categoryCountRow is the temporary projection for one grouped category/count
// aggregate.
type categoryCountRow struct {
	Category string `json:"category"`
	Count    int    `json:"count"`
}

// PlayerHonors returns every honor definition with the player's unlock status.
func (s *serviceImpl) PlayerHonors(ctx context.Context, playerID int64) ([]*PlayerHonorItem, error) {
	defs := make([]*entitymodel.HonorDef, 0)
	err := dao.HonorDef.Ctx(ctx).
		OrderAsc(dao.HonorDef.Columns().Sort).
		OrderDesc(dao.HonorDef.Columns().Id).
		Scan(&defs)
	if err != nil {
		return nil, bizerr.WrapCode(err, CodeHonorQueryFailed)
	}
	if len(defs) == 0 {
		return []*PlayerHonorItem{}, nil
	}

	progress, err := s.aggregateProgress(ctx, playerID)
	if err != nil {
		return nil, err
	}

	items := make([]*PlayerHonorItem, 0, len(defs))
	for _, def := range defs {
		items = append(items, &PlayerHonorItem{
			Id:         def.Id,
			HonorType:  def.HonorType,
			Code:       def.Code,
			Name:       def.Name,
			UnlockType: def.UnlockType,
			Threshold:  def.Threshold,
			Category:   def.Category,
			ImagePath:  def.ImagePath,
			Unlocked:   progress.unlocked(UnlockType(def.UnlockType), def.Threshold, def.Category),
		})
	}
	return items, nil
}

// aggregateProgress pre-aggregates all counts the unlock rules need for playerID
// in a fixed number of batched queries: the feeding count, the activation count,
// the player's collected-card counts per category and the active-card totals per
// category. A non-positive playerID yields zeroed player counts while the global
// active-card totals are still loaded so completion rules stay well defined.
func (s *serviceImpl) aggregateProgress(ctx context.Context, playerID int64) (*playerProgress, error) {
	progress := &playerProgress{
		collectedByCategory: map[string]int{},
		activeByCategory:    map[string]int{},
	}

	activeByCategory, activeTotal, err := s.activeCardsByCategory(ctx)
	if err != nil {
		return nil, err
	}
	progress.activeByCategory = activeByCategory
	progress.activeTotal = activeTotal

	if playerID <= 0 {
		return progress, nil
	}

	feedCount, err := s.playerFeedCount(ctx, playerID)
	if err != nil {
		return nil, err
	}
	progress.feedCount = feedCount

	activationCount, err := s.playerActivationCount(ctx, playerID)
	if err != nil {
		return nil, err
	}
	progress.activationCount = activationCount

	collectedByCategory, collectedTotal, err := s.playerCollectedByCategory(ctx, playerID)
	if err != nil {
		return nil, err
	}
	progress.collectedByCategory = collectedByCategory
	progress.collectedTotal = collectedTotal

	return progress, nil
}

// playerFeedCount returns the number of feeding records the player has.
func (s *serviceImpl) playerFeedCount(ctx context.Context, playerID int64) (int, error) {
	count, err := dao.Feeding.Ctx(ctx).
		Where(dao.Feeding.Columns().UserId, playerID).
		Count()
	if err != nil {
		return 0, bizerr.WrapCode(err, CodeHonorQueryFailed)
	}
	return count, nil
}

// playerActivationCount returns the number of activation records the player has.
func (s *serviceImpl) playerActivationCount(ctx context.Context, playerID int64) (int, error) {
	count, err := dao.Activation.Ctx(ctx).
		Where(dao.Activation.Columns().UserId, playerID).
		Count()
	if err != nil {
		return 0, bizerr.WrapCode(err, CodeHonorQueryFailed)
	}
	return count, nil
}

// activeCardsByCategory returns the per-category and total counts of active main
// cards in two database-side aggregates. The per-category counts are grouped on
// the database side; the total is summed from them so no extra query is needed.
func (s *serviceImpl) activeCardsByCategory(ctx context.Context) (map[string]int, int, error) {
	rows := make([]*categoryCountRow, 0)
	err := dao.Card.Ctx(ctx).
		Fields(dao.Card.Columns().Category, "COUNT(*) AS count").
		Group(dao.Card.Columns().Category).
		Scan(&rows)
	if err != nil {
		return nil, 0, bizerr.WrapCode(err, CodeHonorQueryFailed)
	}
	byCategory := make(map[string]int, len(rows))
	total := 0
	for _, row := range rows {
		byCategory[row.Category] = row.Count
		total += row.Count
	}
	return byCategory, total, nil
}

// playerCollectedByCategory returns the per-category and total counts of active
// cards the player has collected. A collected card is the main card of a cattle
// the player activated; the activation table is joined to the card table on the
// cattle ID, restricted to the player and grouped by card category on the database
// side. The total is summed from the grouped counts.
func (s *serviceImpl) playerCollectedByCategory(ctx context.Context, playerID int64) (map[string]int, int, error) {
	rows := make([]*categoryCountRow, 0)
	err := dao.Activation.Ctx(ctx).
		As("a").
		InnerJoin(
			cardTable+" AS c",
			"c."+dao.Card.Columns().NiuId+" = a."+dao.Activation.Columns().NiuId+
				" AND c."+dao.Card.Columns().DeletedAt+" IS NULL",
		).
		Where("a."+dao.Activation.Columns().UserId, playerID).
		Fields(
			"c."+dao.Card.Columns().Category+" AS category",
			"COUNT(*) AS count",
		).
		Group("c." + dao.Card.Columns().Category).
		Scan(&rows)
	if err != nil {
		return nil, 0, bizerr.WrapCode(err, CodeHonorQueryFailed)
	}
	byCategory := make(map[string]int, len(rows))
	total := 0
	for _, row := range rows {
		byCategory[row.Category] = row.Count
		total += row.Count
	}
	return byCategory, total, nil
}

// cardTable is the qualified card table name used to build the inner join in the
// player collected-card aggregation; it mirrors the DAO-owned table name.
const cardTable = "plugin_sicau_niu_card"

// unlocked evaluates one unlock rule in memory against the pre-aggregated counts.
// participation is always unlocked; feed_count and activation_count compare the
// player's count against the threshold; category_complete requires the player to
// have collected every active card of the configured category (which must have at
// least one active card); full_complete requires the player to have collected
// every active card overall (which must have at least one active card).
func (p *playerProgress) unlocked(unlockType UnlockType, threshold int, category string) bool {
	switch unlockType {
	case UnlockTypeParticipation:
		return true
	case UnlockTypeFeedCount:
		return threshold > 0 && p.feedCount >= threshold
	case UnlockTypeActivationCount:
		return threshold > 0 && p.activationCount >= threshold
	case UnlockTypeCategoryComplete:
		active := p.activeByCategory[category]
		return active > 0 && p.collectedByCategory[category] >= active
	case UnlockTypeFullComplete:
		return p.activeTotal > 0 && p.collectedTotal >= p.activeTotal
	default:
		return false
	}
}
