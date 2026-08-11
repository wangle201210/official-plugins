package activation

import (
	"context"

	"lina-core/pkg/bizerr"
	"lina-plugin-sicau-niu/backend/internal/dao"
	"lina-plugin-sicau-niu/backend/internal/model/do"
	entitymodel "lina-plugin-sicau-niu/backend/internal/model/entity"
)

type NiuDetailOutput struct {
	Niu   *VisibleNiuItem
	Quote string
	Card  *NiuDetailCard
}

type NiuDetailCard struct {
	ID        int64
	Category  string
	Title     string
	Content   string
	ImagePath string
	Owned     bool
}

// NiuDetail reuses the bounded map projection so location disclosure and
// aggregate semantics stay identical between list and detail endpoints.
func (s *serviceImpl) NiuDetail(ctx context.Context, playerID, niuID int64) (*NiuDetailOutput, error) {
	if niuID <= 0 {
		return nil, bizerr.NewCode(CodeNiuIDRequired)
	}
	items, err := s.VisibleNiu(ctx, playerID)
	if err != nil {
		return nil, err
	}
	var item *VisibleNiuItem
	for _, candidate := range items {
		if candidate.Id == niuID {
			item = candidate
			break
		}
	}
	if item == nil {
		return nil, bizerr.NewCode(CodeNiuNotFound)
	}
	quote, err := s.randomEnabledQuote(ctx)
	if err != nil {
		return nil, err
	}
	output := &NiuDetailOutput{Niu: item, Quote: quote}
	if !item.ActivatedByMe {
		return output, nil
	}
	var card *entitymodel.Card
	if err = dao.Card.Ctx(ctx).Where(do.Card{NiuId: niuID}).Scan(&card); err != nil {
		return nil, bizerr.WrapCode(err, CodeQueryFailed)
	}
	if card != nil {
		output.Card = &NiuDetailCard{ID: card.Id, Category: card.Category, Title: card.Title, Content: card.Content, ImagePath: card.ImagePath, Owned: true}
	}
	return output, nil
}
