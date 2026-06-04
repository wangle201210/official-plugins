// honor_crud.go implements the operator honor-definition CRUD: DB-side paged
// listing, detail, create, update and soft-delete. Mutations enforce
// honor-type/unlock-type enum validation, the threshold/category shape implied by
// the unlock rule and code uniqueness with bounded queries.

package honor

import (
	"context"
	"strings"

	"lina-core/pkg/apitime"
	"lina-core/pkg/bizerr"
	"lina-plugin-sicau-niu/backend/internal/dao"
	"lina-plugin-sicau-niu/backend/internal/model/do"
	entitymodel "lina-plugin-sicau-niu/backend/internal/model/entity"
	cardsvc "lina-plugin-sicau-niu/backend/internal/service/card"
)

// ListInput defines the operator honor-definition list query.
type ListInput struct {
	// Keyword is the optional fuzzy match applied to the honor code or name.
	Keyword string
	// HonorType optionally filters honors by type; empty lists all.
	HonorType string
	// UnlockType optionally filters honors by unlock rule; empty lists all.
	UnlockType string
	// PageNum is the requested page number; defaults to 1 when non-positive.
	PageNum int
	// PageSize is the requested page size; defaults to 10 and is capped at 100.
	PageSize int
}

// ListOutput defines the operator honor-definition list result.
type ListOutput struct {
	// List holds the current page of honor definitions.
	List []*HonorItem
	// Total is the total matched honor count.
	Total int
}

// HonorItem defines one honor-definition row projected for the operator console.
type HonorItem struct {
	Id         int64
	HonorType  string
	Code       string
	Name       string
	UnlockType string
	Threshold  int
	Category   string
	ImagePath  string
	Sort       int
	CreatedAt  *int64
	UpdatedAt  *int64
}

// MutateInput defines the create/update honor-definition input.
type MutateInput struct {
	HonorType  string
	Code       string
	Name       string
	UnlockType string
	Threshold  int
	Category   string
	ImagePath  string
	Sort       int
}

// List returns one DB-side paged honor-definition page ordered by sort then ID.
func (s *serviceImpl) List(ctx context.Context, in *ListInput) (*ListOutput, error) {
	pageNum, pageSize := defaultPageNum, defaultPageSize
	model := dao.HonorDef.Ctx(ctx)
	if in != nil {
		pageNum, pageSize = normalizePagination(in.PageNum, in.PageSize)
		if keyword := strings.TrimSpace(in.Keyword); keyword != "" {
			like := "%" + keyword + "%"
			keywordFilter := model.Builder().
				WhereLike(dao.HonorDef.Columns().Code, like).
				WhereOrLike(dao.HonorDef.Columns().Name, like)
			model = model.Where(keywordFilter)
		}
		if honorType := strings.TrimSpace(in.HonorType); honorType != "" {
			model = model.Where(do.HonorDef{HonorType: honorType})
		}
		if unlockType := strings.TrimSpace(in.UnlockType); unlockType != "" {
			model = model.Where(do.HonorDef{UnlockType: unlockType})
		}
	}

	total, err := model.Count()
	if err != nil {
		return nil, bizerr.WrapCode(err, CodeHonorQueryFailed)
	}

	rows := make([]*entitymodel.HonorDef, 0)
	err = model.
		OrderAsc(dao.HonorDef.Columns().Sort).
		OrderDesc(dao.HonorDef.Columns().Id).
		Page(pageNum, pageSize).
		Scan(&rows)
	if err != nil {
		return nil, bizerr.WrapCode(err, CodeHonorQueryFailed)
	}

	list := make([]*HonorItem, 0, len(rows))
	for _, row := range rows {
		list = append(list, toHonorItem(row))
	}
	return &ListOutput{List: list, Total: total}, nil
}

// Get returns one honor-definition detail by ID.
func (s *serviceImpl) Get(ctx context.Context, id int64) (*HonorItem, error) {
	if id <= 0 {
		return nil, bizerr.NewCode(CodeHonorIDRequired)
	}
	var row *entitymodel.HonorDef
	err := dao.HonorDef.Ctx(ctx).Where(do.HonorDef{Id: id}).Scan(&row)
	if err != nil {
		return nil, bizerr.WrapCode(err, CodeHonorQueryFailed)
	}
	if row == nil {
		return nil, bizerr.NewCode(CodeHonorNotFound)
	}
	return toHonorItem(row), nil
}

// Create inserts one honor definition after validation.
func (s *serviceImpl) Create(ctx context.Context, in *MutateInput) (int64, error) {
	data, err := s.validate(ctx, in, 0)
	if err != nil {
		return 0, err
	}
	id, err := dao.HonorDef.Ctx(ctx).Data(data).InsertAndGetId()
	if err != nil {
		return 0, bizerr.WrapCode(err, CodeHonorWriteFailed)
	}
	return id, nil
}

// Update modifies one honor definition after existence and validation.
func (s *serviceImpl) Update(ctx context.Context, id int64, in *MutateInput) error {
	if id <= 0 {
		return bizerr.NewCode(CodeHonorIDRequired)
	}
	exists, err := s.exists(ctx, id)
	if err != nil {
		return err
	}
	if !exists {
		return bizerr.NewCode(CodeHonorNotFound)
	}

	data, err := s.validate(ctx, in, id)
	if err != nil {
		return err
	}
	_, err = dao.HonorDef.Ctx(ctx).Where(do.HonorDef{Id: id}).Data(data).Update()
	if err != nil {
		return bizerr.WrapCode(err, CodeHonorWriteFailed)
	}
	return nil
}

// Delete soft-deletes one honor definition after existence validation.
func (s *serviceImpl) Delete(ctx context.Context, id int64) error {
	if id <= 0 {
		return bizerr.NewCode(CodeHonorIDRequired)
	}
	exists, err := s.exists(ctx, id)
	if err != nil {
		return err
	}
	if !exists {
		return bizerr.NewCode(CodeHonorNotFound)
	}
	if _, err = dao.HonorDef.Ctx(ctx).Where(do.HonorDef{Id: id}).Delete(); err != nil {
		return bizerr.WrapCode(err, CodeHonorWriteFailed)
	}
	return nil
}

// exists reports whether an active honor definition with id exists.
func (s *serviceImpl) exists(ctx context.Context, id int64) (bool, error) {
	if id <= 0 {
		return false, nil
	}
	count, err := dao.HonorDef.Ctx(ctx).Where(do.HonorDef{Id: id}).Count()
	if err != nil {
		return false, bizerr.WrapCode(err, CodeHonorQueryFailed)
	}
	return count > 0, nil
}

// validate validates the mutate input and builds the persisted DO. It enforces
// code presence/uniqueness, name presence, honor-type/unlock-type enum rules and
// the threshold/category shape implied by the unlock rule. excludeID is the honor
// excluded from the code-uniqueness check on update so an unchanged code does not
// collide with itself.
func (s *serviceImpl) validate(ctx context.Context, in *MutateInput, excludeID int64) (do.HonorDef, error) {
	if in == nil {
		return do.HonorDef{}, bizerr.NewCode(CodeHonorCodeRequired)
	}
	code := strings.TrimSpace(in.Code)
	if code == "" {
		return do.HonorDef{}, bizerr.NewCode(CodeHonorCodeRequired)
	}
	name := strings.TrimSpace(in.Name)
	if name == "" {
		return do.HonorDef{}, bizerr.NewCode(CodeHonorNameRequired)
	}

	honorType := HonorType(strings.TrimSpace(in.HonorType))
	if !honorType.valid() {
		return do.HonorDef{}, bizerr.NewCode(CodeHonorTypeInvalid)
	}
	unlockType := UnlockType(strings.TrimSpace(in.UnlockType))
	if !unlockType.valid() {
		return do.HonorDef{}, bizerr.NewCode(CodeHonorUnlockTypeInvalid)
	}

	threshold, category, err := validateUnlockShape(unlockType, in)
	if err != nil {
		return do.HonorDef{}, err
	}

	taken, err := s.codeTaken(ctx, code, excludeID)
	if err != nil {
		return do.HonorDef{}, err
	}
	if taken {
		return do.HonorDef{}, bizerr.NewCode(CodeHonorCodeExists)
	}

	return do.HonorDef{
		HonorType:  honorType.String(),
		Code:       code,
		Name:       name,
		UnlockType: unlockType.String(),
		Threshold:  threshold,
		Category:   category,
		ImagePath:  strings.TrimSpace(in.ImagePath),
		Sort:       in.Sort,
	}, nil
}

// validateUnlockShape validates the threshold/category shape implied by the
// unlock rule and returns the threshold and category to persist. Count-based
// rules require a positive threshold and clear the category; the category-complete
// rule requires a valid card category and clears the threshold; the remaining
// rules clear both.
func validateUnlockShape(unlockType UnlockType, in *MutateInput) (threshold int, category string, err error) {
	if unlockType.requiresThreshold() {
		if in.Threshold <= 0 {
			return 0, "", bizerr.NewCode(CodeHonorThresholdInvalid)
		}
		return in.Threshold, "", nil
	}
	if unlockType.requiresCategory() {
		cat := strings.TrimSpace(in.Category)
		if cat == "" || !cardsvc.ValidCategory(cat) {
			return 0, "", bizerr.NewCode(CodeHonorCategoryInvalid)
		}
		return 0, cat, nil
	}
	return 0, "", nil
}

// codeTaken reports whether code is owned by another active honor definition. When
// excludeID is positive that honor is excluded so an unchanged code on update does
// not collide with itself.
func (s *serviceImpl) codeTaken(ctx context.Context, code string, excludeID int64) (bool, error) {
	model := dao.HonorDef.Ctx(ctx).Where(do.HonorDef{Code: code})
	if excludeID > 0 {
		model = model.WhereNot(dao.HonorDef.Columns().Id, excludeID)
	}
	count, err := model.Count()
	if err != nil {
		return false, bizerr.WrapCode(err, CodeHonorQueryFailed)
	}
	return count > 0, nil
}

// toHonorItem projects one honor-definition entity to its operator item.
func toHonorItem(row *entitymodel.HonorDef) *HonorItem {
	return &HonorItem{
		Id:         row.Id,
		HonorType:  row.HonorType,
		Code:       row.Code,
		Name:       row.Name,
		UnlockType: row.UnlockType,
		Threshold:  row.Threshold,
		Category:   row.Category,
		ImagePath:  row.ImagePath,
		Sort:       row.Sort,
		CreatedAt:  apitime.Milli(row.CreatedAt),
		UpdatedAt:  apitime.Milli(row.UpdatedAt),
	}
}
