// wall_highlights.go implements the public campus-history highlights: a bounded
// sample of cattle history cards and enabled campus-history quotes. The cattle
// names for the sampled cards are batch-assembled in one projected query to avoid
// N+1. Both samples are capped so the public endpoint is always bounded.

package wall

import (
	"context"

	"lina-core/pkg/bizerr"
	"lina-plugin-sicau-niu/backend/internal/dao"
	entitymodel "lina-plugin-sicau-niu/backend/internal/model/entity"
)

// Highlights is the public campus-history highlights result.
type Highlights struct {
	// Cards is the bounded sample of campus-history cards.
	Cards []*HighlightCard
	// Quotes is the bounded sample of enabled campus-history quotes.
	Quotes []*HighlightQuote
}

// HighlightCard is one campus-history card on the public wall.
type HighlightCard struct {
	Id        int64
	NiuId     int64
	NiuName   string
	Category  string
	Title     string
	Content   string
	ImagePath string
}

// HighlightQuote is one enabled campus-history quote on the public wall.
type HighlightQuote struct {
	Id      int64
	Content string
}

// Highlights returns the bounded campus-history card and quote samples.
func (s *serviceImpl) Highlights(ctx context.Context) (*Highlights, error) {
	cards, err := s.sampleCards(ctx)
	if err != nil {
		return nil, err
	}
	quotes, err := s.sampleQuotes(ctx)
	if err != nil {
		return nil, err
	}
	return &Highlights{Cards: cards, Quotes: quotes}, nil
}

// sampleCards reads the latest campus-history cards capped at highlightCardLimit
// and batch-assembles their cattle names in one query.
func (s *serviceImpl) sampleCards(ctx context.Context) ([]*HighlightCard, error) {
	rows := make([]*entitymodel.Card, 0, highlightCardLimit)
	err := dao.Card.Ctx(ctx).
		Order(dao.Card.Columns().Id + " DESC").
		Limit(highlightCardLimit).
		Scan(&rows)
	if err != nil {
		return nil, bizerr.WrapCode(err, CodeWallQueryFailed)
	}
	if len(rows) == 0 {
		return []*HighlightCard{}, nil
	}
	niuIDs := make([]int64, 0, len(rows))
	for _, row := range rows {
		niuIDs = append(niuIDs, row.NiuId)
	}
	niu, err := s.batchNiu(ctx, niuIDs)
	if err != nil {
		return nil, err
	}
	cards := make([]*HighlightCard, 0, len(rows))
	for _, row := range rows {
		card := &HighlightCard{
			Id:        row.Id,
			NiuId:     row.NiuId,
			Category:  row.Category,
			Title:     row.Title,
			Content:   row.Content,
			ImagePath: row.ImagePath,
		}
		if cattle := niu[row.NiuId]; cattle != nil {
			card.NiuName = cattle.Name
		}
		cards = append(cards, card)
	}
	return cards, nil
}

// sampleQuotes reads the latest enabled campus-history quotes capped at
// highlightQuoteLimit.
func (s *serviceImpl) sampleQuotes(ctx context.Context) ([]*HighlightQuote, error) {
	rows := make([]*entitymodel.Quote, 0, highlightQuoteLimit)
	err := dao.Quote.Ctx(ctx).
		Where(dao.Quote.Columns().Enabled, 1).
		Order(dao.Quote.Columns().Id + " DESC").
		Limit(highlightQuoteLimit).
		Scan(&rows)
	if err != nil {
		return nil, bizerr.WrapCode(err, CodeWallQueryFailed)
	}
	quotes := make([]*HighlightQuote, 0, len(rows))
	for _, row := range rows {
		quotes = append(quotes, &HighlightQuote{Id: row.Id, Content: row.Content})
	}
	return quotes, nil
}
