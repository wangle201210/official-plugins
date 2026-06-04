// college_crud.go implements college dictionary listing, creation, update,
// reference-protected deletion, player options and existence checks. All queries
// run DB-side filtering, ordering and pagination; name uniqueness and
// delete-time reference protection are enforced with bounded count queries to
// avoid loading full collections.

package college

import (
	"context"
	"strings"

	"lina-core/pkg/apitime"
	"lina-core/pkg/bizerr"
	"lina-plugin-sicau-niu/backend/internal/dao"
	"lina-plugin-sicau-niu/backend/internal/model/do"
	entitymodel "lina-plugin-sicau-niu/backend/internal/model/entity"
)

// College listing paging defaults bound the operator console page size.
const (
	defaultPageNum  = 1
	defaultPageSize = 10
	maxPageSize     = 100
)

// ListInput defines the college list query.
type ListInput struct {
	// Keyword is the optional fuzzy match applied to the college name.
	Keyword string
	// PageNum is the requested page number; defaults to 1 when non-positive.
	PageNum int
	// PageSize is the requested page size; defaults to 10 and is capped at 100.
	PageSize int
}

// ListOutput defines the college list result.
type ListOutput struct {
	// List holds the current page of colleges.
	List []*Item
	// Total is the total matched college count.
	Total int
}

// Item defines one college row projected for the operator console.
type Item struct {
	// Id is the college ID.
	Id int64
	// Name is the college name.
	Name string
	// Sort is the display sort order.
	Sort int
	// CreatedAt is the creation time as a Unix timestamp in milliseconds.
	CreatedAt *int64
	// UpdatedAt is the update time as a Unix timestamp in milliseconds.
	UpdatedAt *int64
}

// OptionItem defines one college option for player identity selection.
type OptionItem struct {
	// Id is the college ID.
	Id int64
	// Name is the college name.
	Name string
}

// MutateInput defines the create/update college input.
type MutateInput struct {
	// Name is the required college name.
	Name string
	// Sort is the display sort order, smaller first.
	Sort int
}

// List returns one DB-side paged, sort-ordered college page.
func (s *serviceImpl) List(ctx context.Context, in *ListInput) (*ListOutput, error) {
	pageNum, pageSize := normalizePagination(in)
	model := dao.College.Ctx(ctx)
	if in != nil {
		keyword := strings.TrimSpace(in.Keyword)
		if keyword != "" {
			model = model.WhereLike(dao.College.Columns().Name, "%"+keyword+"%")
		}
	}

	total, err := model.Count()
	if err != nil {
		return nil, bizerr.WrapCode(err, CodeCollegeQueryFailed)
	}

	rows := make([]*entitymodel.College, 0)
	err = model.
		OrderAsc(dao.College.Columns().Sort).
		OrderDesc(dao.College.Columns().Id).
		Page(pageNum, pageSize).
		Scan(&rows)
	if err != nil {
		return nil, bizerr.WrapCode(err, CodeCollegeQueryFailed)
	}

	list := make([]*Item, 0, len(rows))
	for _, row := range rows {
		list = append(list, &Item{
			Id:        row.Id,
			Name:      row.Name,
			Sort:      row.Sort,
			CreatedAt: apitime.Milli(row.CreatedAt),
			UpdatedAt: apitime.Milli(row.UpdatedAt),
		})
	}
	return &ListOutput{List: list, Total: total}, nil
}

// Create inserts one college after validation.
func (s *serviceImpl) Create(ctx context.Context, in *MutateInput) (int64, error) {
	name, err := validateName(in)
	if err != nil {
		return 0, err
	}

	taken, err := s.nameTaken(ctx, name, 0)
	if err != nil {
		return 0, err
	}
	if taken {
		return 0, bizerr.NewCode(CodeCollegeNameExists)
	}

	id, err := dao.College.Ctx(ctx).Data(do.College{
		Name: name,
		Sort: in.Sort,
	}).InsertAndGetId()
	if err != nil {
		return 0, bizerr.WrapCode(err, CodeCollegeWriteFailed)
	}
	return id, nil
}

// Update modifies one college after existence and uniqueness validation.
func (s *serviceImpl) Update(ctx context.Context, id int64, in *MutateInput) error {
	if id <= 0 {
		return bizerr.NewCode(CodeCollegeIDRequired)
	}
	name, err := validateName(in)
	if err != nil {
		return err
	}

	exists, err := s.Exists(ctx, id)
	if err != nil {
		return err
	}
	if !exists {
		return bizerr.NewCode(CodeCollegeNotFound)
	}

	taken, err := s.nameTaken(ctx, name, id)
	if err != nil {
		return err
	}
	if taken {
		return bizerr.NewCode(CodeCollegeNameExists)
	}

	_, err = dao.College.Ctx(ctx).
		Where(do.College{Id: id}).
		Data(do.College{Name: name, Sort: in.Sort}).
		Update()
	if err != nil {
		return bizerr.WrapCode(err, CodeCollegeWriteFailed)
	}
	return nil
}

// Delete soft-deletes one college after verifying it is not referenced.
func (s *serviceImpl) Delete(ctx context.Context, id int64) error {
	if id <= 0 {
		return bizerr.NewCode(CodeCollegeIDRequired)
	}

	exists, err := s.Exists(ctx, id)
	if err != nil {
		return err
	}
	if !exists {
		return bizerr.NewCode(CodeCollegeNotFound)
	}

	referenced, err := s.referencedByPlayer(ctx, id)
	if err != nil {
		return err
	}
	if referenced {
		return bizerr.NewCode(CodeCollegeReferenced)
	}

	_, err = dao.College.Ctx(ctx).Where(do.College{Id: id}).Delete()
	if err != nil {
		return bizerr.WrapCode(err, CodeCollegeWriteFailed)
	}
	return nil
}

// Options returns the bounded, sort-ordered college options in one query.
func (s *serviceImpl) Options(ctx context.Context) ([]*OptionItem, error) {
	rows := make([]*entitymodel.College, 0)
	err := dao.College.Ctx(ctx).
		Fields(dao.College.Columns().Id, dao.College.Columns().Name).
		OrderAsc(dao.College.Columns().Sort).
		OrderAsc(dao.College.Columns().Id).
		Scan(&rows)
	if err != nil {
		return nil, bizerr.WrapCode(err, CodeCollegeQueryFailed)
	}

	options := make([]*OptionItem, 0, len(rows))
	for _, row := range rows {
		options = append(options, &OptionItem{Id: row.Id, Name: row.Name})
	}
	return options, nil
}

// Exists reports whether an active college with id exists.
func (s *serviceImpl) Exists(ctx context.Context, id int64) (bool, error) {
	if id <= 0 {
		return false, nil
	}
	count, err := dao.College.Ctx(ctx).Where(do.College{Id: id}).Count()
	if err != nil {
		return false, bizerr.WrapCode(err, CodeCollegeQueryFailed)
	}
	return count > 0, nil
}

// nameTaken reports whether name is owned by another active college. When
// excludeID is positive that college is excluded so an unchanged name on update
// does not collide with itself.
func (s *serviceImpl) nameTaken(ctx context.Context, name string, excludeID int64) (bool, error) {
	model := dao.College.Ctx(ctx).Where(do.College{Name: name})
	if excludeID > 0 {
		model = model.WhereNot(dao.College.Columns().Id, excludeID)
	}
	count, err := model.Count()
	if err != nil {
		return false, bizerr.WrapCode(err, CodeCollegeQueryFailed)
	}
	return count > 0, nil
}

// referencedByPlayer reports whether at least one active player selected the
// college. It uses a bounded count query rather than loading player rows.
func (s *serviceImpl) referencedByPlayer(ctx context.Context, id int64) (bool, error) {
	count, err := dao.User.Ctx(ctx).
		Where(dao.User.Columns().CollegeId, id).
		Count()
	if err != nil {
		return false, bizerr.WrapCode(err, CodeCollegeQueryFailed)
	}
	return count > 0, nil
}

// validateName trims and validates the required college name.
func validateName(in *MutateInput) (string, error) {
	if in == nil {
		return "", bizerr.NewCode(CodeCollegeNameRequired)
	}
	name := strings.TrimSpace(in.Name)
	if name == "" {
		return "", bizerr.NewCode(CodeCollegeNameRequired)
	}
	return name, nil
}

// normalizePagination applies paging defaults and the max page-size cap.
func normalizePagination(in *ListInput) (int, int) {
	if in == nil {
		return defaultPageNum, defaultPageSize
	}
	pageNum := in.PageNum
	if pageNum <= 0 {
		pageNum = defaultPageNum
	}
	pageSize := in.PageSize
	if pageSize <= 0 {
		pageSize = defaultPageSize
	}
	if pageSize > maxPageSize {
		pageSize = maxPageSize
	}
	return pageNum, pageSize
}
