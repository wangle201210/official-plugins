// activation_collection.go implements the player's personal card collection
// (图鉴). It loads a bounded catalog, then batch-loads the player's ownership set
// and cattle codes. Locked entries retain category information for progress
// display but never expose card-face content. The fixed three-query assembly
// avoids N+1 while preserving player isolation.

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

// collectionCap bounds the catalog returned to one mini-program page.
const collectionCap = 500

// CollectionItem defines one collected card with its owning cattle reference.
type CollectionItem struct {
	// NiuId is the owning cattle ID.
	NiuId int64
	// NiuCode is the owning cattle serial code.
	NiuCode string
	// Category is the card category string.
	Category string
	// Owned reports whether the current player owns the card.
	Owned bool
	// Title is the card title; empty while locked.
	Title string
	// Content is the card content text; empty while locked.
	Content string
	// ImagePath is the card image storage path; empty while locked or when none.
	ImagePath string
}

// Collection returns the bounded card catalog and playerID's ownership state,
// optionally filtered by category and ordered by card ID.
func (s *serviceImpl) Collection(ctx context.Context, playerID int64, category string) ([]*CollectionItem, error) {
	if playerID <= 0 {
		return []*CollectionItem{}, nil
	}
	categoryFilter := strings.TrimSpace(category)
	if categoryFilter != "" && !cardsvc.ValidCategory(categoryFilter) {
		return nil, bizerr.NewCode(CodeCategoryInvalid)
	}

	cards, err := s.collectionCards(ctx, categoryFilter)
	if err != nil {
		return nil, err
	}
	if len(cards) == 0 {
		return []*CollectionItem{}, nil
	}

	niuIDs := make([]int64, 0, len(cards))
	for _, card := range cards {
		niuIDs = append(niuIDs, card.NiuId)
	}
	ownedNiuIDs, err := s.ownedNiuIDs(ctx, playerID, niuIDs)
	if err != nil {
		return nil, err
	}
	niuCodes, err := s.batchNiuCodes(ctx, niuIDs)
	if err != nil {
		return nil, err
	}

	list := make([]*CollectionItem, 0, len(cards))
	for _, card := range cards {
		_, owned := ownedNiuIDs[card.NiuId]
		item := &CollectionItem{
			NiuId:    card.NiuId,
			NiuCode:  niuCodes[card.NiuId],
			Category: card.Category,
			Owned:    owned,
		}
		if owned {
			item.Title = card.Title
			item.Content = card.Content
			item.ImagePath = card.ImagePath
		}
		list = append(list, item)
	}
	return list, nil
}

// collectionCards returns the bounded card catalog in stable card-ID order.
func (s *serviceImpl) collectionCards(ctx context.Context, category string) ([]*entitymodel.Card, error) {
	model := dao.Card.Ctx(ctx)
	if category != "" {
		model = model.Where(do.Card{Category: category})
	}
	rows := make([]*entitymodel.Card, 0)
	err := model.
		Fields(
			dao.Card.Columns().Id,
			dao.Card.Columns().NiuId,
			dao.Card.Columns().Category,
			dao.Card.Columns().Title,
			dao.Card.Columns().Content,
			dao.Card.Columns().ImagePath,
		).
		OrderAsc(dao.Card.Columns().Id).
		Limit(collectionCap).
		Scan(&rows)
	if err != nil {
		return nil, bizerr.WrapCode(err, CodeQueryFailed)
	}
	return rows, nil
}

// ownedNiuIDs returns the cattle IDs playerID has activated, restricted to the
// bounded catalog IDs so the query remains bounded by collectionCap.
func (s *serviceImpl) ownedNiuIDs(ctx context.Context, playerID int64, niuIDs []int64) (map[int64]struct{}, error) {
	rows := make([]*entitymodel.Activation, 0, len(niuIDs))
	err := dao.Activation.Ctx(ctx).
		Fields(dao.Activation.Columns().NiuId).
		Where(do.Activation{UserId: playerID}).
		WhereIn(dao.Activation.Columns().NiuId, niuIDs).
		Scan(&rows)
	if err != nil {
		return nil, bizerr.WrapCode(err, CodeQueryFailed)
	}
	owned := make(map[int64]struct{}, len(rows))
	for _, row := range rows {
		owned[row.NiuId] = struct{}{}
	}
	return owned, nil
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
