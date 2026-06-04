// card_quote.go implements quote listing, creation, update and deletion. Listing
// runs DB-side filtering/sorting/pagination. Mutations enforce non-empty content;
// the enabled flag defaults to 1 on creation and stays unchanged when omitted on
// update.

package card

import (
	"context"
	"strings"

	"lina-core/pkg/apitime"
	"lina-core/pkg/bizerr"
	"lina-plugin-sicau-niu/backend/internal/dao"
	"lina-plugin-sicau-niu/backend/internal/model/do"
	entitymodel "lina-plugin-sicau-niu/backend/internal/model/entity"
)

// quoteEnabledDefault is the enabled flag applied to a new quote when the caller
// omits it, so a freshly created quote participates in random playback.
const quoteEnabledDefault = 1

// ListQuoteInput defines the operator quote list query.
type ListQuoteInput struct {
	// Keyword is the optional fuzzy match applied to the quote content.
	Keyword string
	// PageNum is the requested page number; defaults to 1 when non-positive.
	PageNum int
	// PageSize is the requested page size; defaults to 10 and is capped at 100.
	PageSize int
}

// ListQuoteOutput defines the operator quote list result.
type ListQuoteOutput struct {
	// List holds the current page of quotes.
	List []*QuoteItem
	// Total is the total matched quote count.
	Total int
}

// QuoteItem defines one quote row projected for the operator console.
type QuoteItem struct {
	Id        int64
	Content   string
	Enabled   int
	CreatedAt *int64
	UpdatedAt *int64
}

// QuoteMutateInput defines the create/update quote input. Enabled is optional: it
// defaults to 1 on creation and is left unchanged on update when nil.
type QuoteMutateInput struct {
	Content string
	Enabled *int
}

// ListQuote returns one DB-side paged quote page.
func (s *serviceImpl) ListQuote(ctx context.Context, in *ListQuoteInput) (*ListQuoteOutput, error) {
	pageNum, pageSize := defaultPageNum, defaultPageSize
	model := dao.Quote.Ctx(ctx)
	if in != nil {
		pageNum, pageSize = normalizePagination(in.PageNum, in.PageSize)
		if keyword := strings.TrimSpace(in.Keyword); keyword != "" {
			model = model.WhereLike(dao.Quote.Columns().Content, "%"+keyword+"%")
		}
	}

	total, err := model.Count()
	if err != nil {
		return nil, bizerr.WrapCode(err, CodeQuoteQueryFailed)
	}

	rows := make([]*entitymodel.Quote, 0)
	err = model.
		OrderDesc(dao.Quote.Columns().Id).
		Page(pageNum, pageSize).
		Scan(&rows)
	if err != nil {
		return nil, bizerr.WrapCode(err, CodeQuoteQueryFailed)
	}

	list := make([]*QuoteItem, 0, len(rows))
	for _, row := range rows {
		list = append(list, &QuoteItem{
			Id:        row.Id,
			Content:   row.Content,
			Enabled:   row.Enabled,
			CreatedAt: apitime.Milli(row.CreatedAt),
			UpdatedAt: apitime.Milli(row.UpdatedAt),
		})
	}
	return &ListQuoteOutput{List: list, Total: total}, nil
}

// CreateQuote inserts one quote after non-empty content validation. Enabled
// defaults to 1 when nil.
func (s *serviceImpl) CreateQuote(ctx context.Context, in *QuoteMutateInput) (int64, error) {
	content, err := validateQuoteContent(in)
	if err != nil {
		return 0, err
	}

	enabled := quoteEnabledDefault
	if in.Enabled != nil {
		enabled = normalizeEnabled(*in.Enabled)
	}

	id, err := dao.Quote.Ctx(ctx).Data(do.Quote{
		Content: content,
		Enabled: enabled,
	}).InsertAndGetId()
	if err != nil {
		return 0, bizerr.WrapCode(err, CodeQuoteWriteFailed)
	}
	return id, nil
}

// UpdateQuote modifies one quote after existence and non-empty content
// validation. A nil Enabled leaves the flag unchanged.
func (s *serviceImpl) UpdateQuote(ctx context.Context, id int64, in *QuoteMutateInput) error {
	if id <= 0 {
		return bizerr.NewCode(CodeQuoteIDRequired)
	}
	content, err := validateQuoteContent(in)
	if err != nil {
		return err
	}

	exists, err := s.quoteExists(ctx, id)
	if err != nil {
		return err
	}
	if !exists {
		return bizerr.NewCode(CodeQuoteNotFound)
	}

	data := do.Quote{Content: content}
	if in.Enabled != nil {
		data.Enabled = normalizeEnabled(*in.Enabled)
	}

	_, err = dao.Quote.Ctx(ctx).Where(do.Quote{Id: id}).Data(data).Update()
	if err != nil {
		return bizerr.WrapCode(err, CodeQuoteWriteFailed)
	}
	return nil
}

// DeleteQuote soft-deletes one quote.
func (s *serviceImpl) DeleteQuote(ctx context.Context, id int64) error {
	if id <= 0 {
		return bizerr.NewCode(CodeQuoteIDRequired)
	}
	exists, err := s.quoteExists(ctx, id)
	if err != nil {
		return err
	}
	if !exists {
		return bizerr.NewCode(CodeQuoteNotFound)
	}

	_, err = dao.Quote.Ctx(ctx).Where(do.Quote{Id: id}).Delete()
	if err != nil {
		return bizerr.WrapCode(err, CodeQuoteWriteFailed)
	}
	return nil
}

// quoteExists reports whether an active quote with id exists.
func (s *serviceImpl) quoteExists(ctx context.Context, id int64) (bool, error) {
	count, err := dao.Quote.Ctx(ctx).Where(do.Quote{Id: id}).Count()
	if err != nil {
		return false, bizerr.WrapCode(err, CodeQuoteQueryFailed)
	}
	return count > 0, nil
}

// validateQuoteContent trims and validates the required quote content.
func validateQuoteContent(in *QuoteMutateInput) (string, error) {
	if in == nil {
		return "", bizerr.NewCode(CodeQuoteContentRequired)
	}
	content := strings.TrimSpace(in.Content)
	if content == "" {
		return "", bizerr.NewCode(CodeQuoteContentRequired)
	}
	return content, nil
}

// normalizeEnabled clamps an arbitrary enabled input to the persisted 0/1 flag,
// treating any non-zero value as enabled.
func normalizeEnabled(value int) int {
	if value == 0 {
		return 0
	}
	return 1
}
