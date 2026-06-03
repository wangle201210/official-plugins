// card_card.go implements card listing, detail, creation, update and deletion.
// Listing runs DB-side filtering/sorting/pagination and then batch-assembles the
// owning-cattle code and name in one bounded query to avoid N+1. Mutations
// enforce category-enum validation, owning-cattle existence and the
// one-card-per-cattle constraint with bounded queries.

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

// ListCardInput defines the operator card list query.
type ListCardInput struct {
	// Keyword is the optional fuzzy match applied to the card title.
	Keyword string
	// Category optionally filters cards by category; empty lists all.
	Category string
	// PageNum is the requested page number; defaults to 1 when non-positive.
	PageNum int
	// PageSize is the requested page size; defaults to 10 and is capped at 100.
	PageSize int
}

// ListCardOutput defines the operator card list result.
type ListCardOutput struct {
	// List holds the current page of cards.
	List []*CardItem
	// Total is the total matched card count.
	Total int
}

// CardItem defines one card row projected for the operator console, including the
// batch-assembled owning-cattle code and name.
type CardItem struct {
	Id        int64
	NiuId     int64
	NiuCode   string
	NiuName   string
	Category  string
	Title     string
	Content   string
	ImagePath string
	CreatedAt *int64
	UpdatedAt *int64
}

// CardMutateInput defines the create/update card input.
type CardMutateInput struct {
	NiuId     int64
	Category  string
	Title     string
	Content   string
	ImagePath string
}

// ListCard returns one DB-side paged card page with batch-assembled cattle info.
func (s *serviceImpl) ListCard(ctx context.Context, in *ListCardInput) (*ListCardOutput, error) {
	pageNum, pageSize := defaultPageNum, defaultPageSize
	model := dao.Card.Ctx(ctx)
	if in != nil {
		pageNum, pageSize = normalizePagination(in.PageNum, in.PageSize)
		if keyword := strings.TrimSpace(in.Keyword); keyword != "" {
			model = model.WhereLike(dao.Card.Columns().Title, "%"+keyword+"%")
		}
		if category := strings.TrimSpace(in.Category); category != "" {
			model = model.Where(do.Card{Category: category})
		}
	}

	total, err := model.Count()
	if err != nil {
		return nil, bizerr.WrapCode(err, CodeCardQueryFailed)
	}

	rows := make([]*entitymodel.Card, 0)
	err = model.
		OrderDesc(dao.Card.Columns().Id).
		Page(pageNum, pageSize).
		Scan(&rows)
	if err != nil {
		return nil, bizerr.WrapCode(err, CodeCardQueryFailed)
	}

	list, err := s.assembleCardItems(ctx, rows)
	if err != nil {
		return nil, err
	}
	return &ListCardOutput{List: list, Total: total}, nil
}

// GetCard returns one card detail with its batch-assembled cattle info.
func (s *serviceImpl) GetCard(ctx context.Context, id int64) (*CardItem, error) {
	if id <= 0 {
		return nil, bizerr.NewCode(CodeCardIDRequired)
	}
	var row *entitymodel.Card
	err := dao.Card.Ctx(ctx).Where(do.Card{Id: id}).Scan(&row)
	if err != nil {
		return nil, bizerr.WrapCode(err, CodeCardQueryFailed)
	}
	if row == nil {
		return nil, bizerr.NewCode(CodeCardNotFound)
	}
	items, err := s.assembleCardItems(ctx, []*entitymodel.Card{row})
	if err != nil {
		return nil, err
	}
	return items[0], nil
}

// CreateCard inserts one card after validation.
func (s *serviceImpl) CreateCard(ctx context.Context, in *CardMutateInput) (int64, error) {
	data, err := s.validateCard(ctx, in, 0)
	if err != nil {
		return 0, err
	}

	id, err := dao.Card.Ctx(ctx).Data(data).InsertAndGetId()
	if err != nil {
		return 0, bizerr.WrapCode(err, CodeCardWriteFailed)
	}
	return id, nil
}

// UpdateCard modifies one card after existence and validation.
func (s *serviceImpl) UpdateCard(ctx context.Context, id int64, in *CardMutateInput) error {
	if id <= 0 {
		return bizerr.NewCode(CodeCardIDRequired)
	}
	exists, err := s.cardExists(ctx, id)
	if err != nil {
		return err
	}
	if !exists {
		return bizerr.NewCode(CodeCardNotFound)
	}

	data, err := s.validateCard(ctx, in, id)
	if err != nil {
		return err
	}

	_, err = dao.Card.Ctx(ctx).Where(do.Card{Id: id}).Data(data).Update()
	if err != nil {
		return bizerr.WrapCode(err, CodeCardWriteFailed)
	}
	return nil
}

// DeleteCard soft-deletes one card, freeing its owning cattle to be bound again.
func (s *serviceImpl) DeleteCard(ctx context.Context, id int64) error {
	if id <= 0 {
		return bizerr.NewCode(CodeCardIDRequired)
	}
	exists, err := s.cardExists(ctx, id)
	if err != nil {
		return err
	}
	if !exists {
		return bizerr.NewCode(CodeCardNotFound)
	}

	_, err = dao.Card.Ctx(ctx).Where(do.Card{Id: id}).Delete()
	if err != nil {
		return bizerr.WrapCode(err, CodeCardWriteFailed)
	}
	return nil
}

// cardExists reports whether an active card with id exists.
func (s *serviceImpl) cardExists(ctx context.Context, id int64) (bool, error) {
	count, err := dao.Card.Ctx(ctx).Where(do.Card{Id: id}).Count()
	if err != nil {
		return false, bizerr.WrapCode(err, CodeCardQueryFailed)
	}
	return count > 0, nil
}

// assembleCardItems projects card rows to operator items, batch-loading the
// owning-cattle code and name in one bounded query keyed by the page's distinct
// cattle IDs to avoid N+1.
func (s *serviceImpl) assembleCardItems(ctx context.Context, rows []*entitymodel.Card) ([]*CardItem, error) {
	ids := make([]int64, 0, len(rows))
	seen := make(map[int64]struct{}, len(rows))
	for _, row := range rows {
		if row.NiuId <= 0 {
			continue
		}
		if _, ok := seen[row.NiuId]; ok {
			continue
		}
		seen[row.NiuId] = struct{}{}
		ids = append(ids, row.NiuId)
	}

	codes := make(map[int64]string, len(ids))
	names := make(map[int64]string, len(ids))
	if len(ids) > 0 {
		niuRows := make([]*entitymodel.Niu, 0, len(ids))
		err := dao.Niu.Ctx(ctx).
			Fields(dao.Niu.Columns().Id, dao.Niu.Columns().Code, dao.Niu.Columns().Name).
			WhereIn(dao.Niu.Columns().Id, ids).
			Scan(&niuRows)
		if err != nil {
			return nil, bizerr.WrapCode(err, CodeCardQueryFailed)
		}
		for _, niuRow := range niuRows {
			codes[niuRow.Id] = niuRow.Code
			names[niuRow.Id] = niuRow.Name
		}
	}

	list := make([]*CardItem, 0, len(rows))
	for _, row := range rows {
		list = append(list, &CardItem{
			Id:        row.Id,
			NiuId:     row.NiuId,
			NiuCode:   codes[row.NiuId],
			NiuName:   names[row.NiuId],
			Category:  row.Category,
			Title:     row.Title,
			Content:   row.Content,
			ImagePath: row.ImagePath,
			CreatedAt: apitime.Milli(row.CreatedAt),
			UpdatedAt: apitime.Milli(row.UpdatedAt),
		})
	}
	return list, nil
}

// validateCard validates the mutate input and builds the persisted DO. It
// enforces category-enum validation, title presence, owning-cattle existence and
// the one-card-per-cattle constraint. excludeID is the card excluded from the
// per-cattle uniqueness check on update so re-saving the same card does not
// collide with itself.
func (s *serviceImpl) validateCard(ctx context.Context, in *CardMutateInput, excludeID int64) (do.Card, error) {
	if in == nil {
		return do.Card{}, bizerr.NewCode(CodeCardNiuRequired)
	}
	if in.NiuId <= 0 {
		return do.Card{}, bizerr.NewCode(CodeCardNiuRequired)
	}
	category := Category(strings.TrimSpace(in.Category))
	if !category.valid() {
		return do.Card{}, bizerr.NewCode(CodeCardCategoryInvalid)
	}
	title := strings.TrimSpace(in.Title)
	if title == "" {
		return do.Card{}, bizerr.NewCode(CodeCardTitleRequired)
	}

	niuExists, err := s.cattleSvc.NiuExists(ctx, in.NiuId)
	if err != nil {
		return do.Card{}, err
	}
	if !niuExists {
		return do.Card{}, bizerr.NewCode(CodeCardNiuInvalid)
	}

	taken, err := s.niuCardTaken(ctx, in.NiuId, excludeID)
	if err != nil {
		return do.Card{}, err
	}
	if taken {
		return do.Card{}, bizerr.NewCode(CodeCardNiuTaken)
	}

	return do.Card{
		NiuId:     in.NiuId,
		Category:  category.String(),
		Title:     title,
		Content:   strings.TrimSpace(in.Content),
		ImagePath: strings.TrimSpace(in.ImagePath),
	}, nil
}

// niuCardTaken reports whether niuID already owns an active card. When excludeID
// is positive that card is excluded so an unchanged binding on update does not
// collide with itself.
func (s *serviceImpl) niuCardTaken(ctx context.Context, niuID int64, excludeID int64) (bool, error) {
	model := dao.Card.Ctx(ctx).Where(do.Card{NiuId: niuID})
	if excludeID > 0 {
		model = model.WhereNot(dao.Card.Columns().Id, excludeID)
	}
	count, err := model.Count()
	if err != nil {
		return false, bizerr.WrapCode(err, CodeCardQueryFailed)
	}
	return count > 0, nil
}
