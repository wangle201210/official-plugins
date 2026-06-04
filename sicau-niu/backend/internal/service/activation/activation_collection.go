// activation_collection.go implements the player's personal card collection
// (图鉴). The collection is derived from the player's own activation records: the
// activated cattle IDs (ordered by activation recency) are collected, then the
// main cards for those cattle are batch-loaded in one WHERE IN query keyed by the
// cattle IDs to avoid N+1, optionally filtered by card category. It is isolated to
// the current player by constraining the activation query to the player ID.

package activation

import (
	"context"
	"strings"

	"lina-core/pkg/bizerr"
	"lina-plugin-sicau-niu/backend/internal/dao"
	"lina-plugin-sicau-niu/backend/internal/model/do"
	entitymodel "lina-plugin-sicau-niu/backend/internal/model/entity"
	cardsvc "lina-plugin-sicau-niu/backend/internal/service/card"
)

// collectionCap bounds the personal collection page. A player can activate at
// most one cattle per day, so the lifetime collection stays small and bounded.
const collectionCap = 500

// CollectionItem defines one collected card with its owning cattle reference.
type CollectionItem struct {
	// NiuId is the owning cattle ID.
	NiuId int64
	// NiuCode is the owning cattle serial code.
	NiuCode string
	// Category is the card category string.
	Category string
	// Title is the card title.
	Title string
	// Content is the card content text.
	Content string
	// ImagePath is the card image storage path; empty when none.
	ImagePath string
}

// Collection returns playerID's personal card collection, optionally filtered by
// category, ordered by activation recency (most recent first).
func (s *serviceImpl) Collection(ctx context.Context, playerID int64, category string) ([]*CollectionItem, error) {
	if playerID <= 0 {
		return []*CollectionItem{}, nil
	}
	categoryFilter := strings.TrimSpace(category)
	if categoryFilter != "" && !cardsvc.ValidCategory(categoryFilter) {
		return nil, bizerr.NewCode(CodeCategoryInvalid)
	}

	orderedNiuIDs, err := s.activatedNiuIDs(ctx, playerID)
	if err != nil {
		return nil, err
	}
	if len(orderedNiuIDs) == 0 {
		return []*CollectionItem{}, nil
	}

	cardsByNiu, err := s.batchCardsByNiu(ctx, orderedNiuIDs, categoryFilter)
	if err != nil {
		return nil, err
	}
	niuCodes, err := s.batchNiuCodes(ctx, orderedNiuIDs)
	if err != nil {
		return nil, err
	}

	list := make([]*CollectionItem, 0, len(orderedNiuIDs))
	for _, niuID := range orderedNiuIDs {
		card, ok := cardsByNiu[niuID]
		if !ok {
			continue
		}
		list = append(list, &CollectionItem{
			NiuId:     niuID,
			NiuCode:   niuCodes[niuID],
			Category:  card.Category,
			Title:     card.Title,
			Content:   card.Content,
			ImagePath: card.ImagePath,
		})
	}
	return list, nil
}

// activatedNiuIDs returns the distinct cattle IDs the player has activated,
// ordered by activation recency (most recent first), in one query.
func (s *serviceImpl) activatedNiuIDs(ctx context.Context, playerID int64) ([]int64, error) {
	records := make([]*entitymodel.Activation, 0)
	err := dao.Activation.Ctx(ctx).
		Fields(dao.Activation.Columns().NiuId).
		Where(dao.Activation.Columns().UserId, playerID).
		OrderDesc(dao.Activation.Columns().Id).
		Limit(collectionCap).
		Scan(&records)
	if err != nil {
		return nil, bizerr.WrapCode(err, CodeQueryFailed)
	}
	seen := make(map[int64]struct{}, len(records))
	ids := make([]int64, 0, len(records))
	for _, record := range records {
		if record.NiuId <= 0 {
			continue
		}
		if _, ok := seen[record.NiuId]; ok {
			continue
		}
		seen[record.NiuId] = struct{}{}
		ids = append(ids, record.NiuId)
	}
	return ids, nil
}

// batchCardsByNiu returns a cattle-ID to main-card map for the activated cattle,
// fetched in one WHERE IN query and optionally narrowed to one category.
func (s *serviceImpl) batchCardsByNiu(
	ctx context.Context,
	niuIDs []int64,
	category string,
) (map[int64]*entitymodel.Card, error) {
	model := dao.Card.Ctx(ctx).WhereIn(dao.Card.Columns().NiuId, niuIDs)
	if category != "" {
		model = model.Where(do.Card{Category: category})
	}
	cardRows := make([]*entitymodel.Card, 0, len(niuIDs))
	if err := model.Scan(&cardRows); err != nil {
		return nil, bizerr.WrapCode(err, CodeQueryFailed)
	}
	cards := make(map[int64]*entitymodel.Card, len(cardRows))
	for _, cardRow := range cardRows {
		cards[cardRow.NiuId] = cardRow
	}
	return cards, nil
}

// batchNiuCodes returns a cattle-ID to serial-code map for the given cattle IDs,
// fetched in one projected WHERE IN query.
func (s *serviceImpl) batchNiuCodes(ctx context.Context, niuIDs []int64) (map[int64]string, error) {
	niuRows := make([]*entitymodel.Niu, 0, len(niuIDs))
	err := dao.Niu.Ctx(ctx).
		Fields(dao.Niu.Columns().Id, dao.Niu.Columns().Code).
		WhereIn(dao.Niu.Columns().Id, niuIDs).
		Scan(&niuRows)
	if err != nil {
		return nil, bizerr.WrapCode(err, CodeQueryFailed)
	}
	codes := make(map[int64]string, len(niuRows))
	for _, niuRow := range niuRows {
		codes[niuRow.Id] = niuRow.Code
	}
	return codes, nil
}
