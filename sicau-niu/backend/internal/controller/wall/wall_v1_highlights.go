// wall_v1_highlights.go implements the public campus-history highlights handler
// and its service-to-DTO projections.

package wall

import (
	"context"

	v1 "lina-plugin-sicau-niu/backend/api/wall/v1"
	wallsvc "lina-plugin-sicau-niu/backend/internal/service/wall"
)

// Highlights returns the public campus-history highlights.
func (c *ControllerV1) Highlights(ctx context.Context, req *v1.HighlightsReq) (res *v1.HighlightsRes, err error) {
	highlights, err := c.wallSvc.Highlights(ctx)
	if err != nil {
		return nil, err
	}
	return &v1.HighlightsRes{
		Cards:  toHighlightCards(highlights.Cards),
		Quotes: toHighlightQuotes(highlights.Quotes),
	}, nil
}

// toHighlightCards projects the campus-history card rows to their response DTOs.
func toHighlightCards(rows []*wallsvc.HighlightCard) []*v1.HighlightCard {
	cards := make([]*v1.HighlightCard, 0, len(rows))
	for _, row := range rows {
		cards = append(cards, &v1.HighlightCard{
			Id:        row.Id,
			NiuId:     row.NiuId,
			NiuName:   row.NiuName,
			Category:  row.Category,
			Title:     row.Title,
			Content:   row.Content,
			ImagePath: row.ImagePath,
		})
	}
	return cards
}

// toHighlightQuotes projects the campus-history quote rows to their response DTOs.
func toHighlightQuotes(rows []*wallsvc.HighlightQuote) []*v1.HighlightQuote {
	quotes := make([]*v1.HighlightQuote, 0, len(rows))
	for _, row := range rows {
		quotes = append(quotes, &v1.HighlightQuote{Id: row.Id, Content: row.Content})
	}
	return quotes
}
