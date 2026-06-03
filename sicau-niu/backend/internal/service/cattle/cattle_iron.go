// cattle_iron.go implements iron-cow identifier listing, creation, update and
// deletion. Listing runs DB-side filtering/sorting/pagination. Mutations enforce
// code presence and uniqueness with bounded queries. The real-time location
// columns (last_lat/last_lng/located_at) are owned by the C4 bonus flow and are
// never written on the operator side.

package cattle

import (
	"context"
	"strings"

	"lina-core/pkg/apitime"
	"lina-core/pkg/bizerr"
	"lina-plugin-sicau-niu/backend/internal/dao"
	"lina-plugin-sicau-niu/backend/internal/model/do"
	entitymodel "lina-plugin-sicau-niu/backend/internal/model/entity"
)

// ListIronInput defines the operator iron-cow list query.
type ListIronInput struct {
	// Keyword is the optional fuzzy match applied to the iron-cow code or name.
	Keyword string
	// PageNum is the requested page number; defaults to 1 when non-positive.
	PageNum int
	// PageSize is the requested page size; defaults to 10 and is capped at 100.
	PageSize int
}

// ListIronOutput defines the operator iron-cow list result.
type ListIronOutput struct {
	// List holds the current page of iron-cows.
	List []*IronItem
	// Total is the total matched iron-cow count.
	Total int
}

// IronItem defines one iron-cow row projected for the operator console.
type IronItem struct {
	Id        int64
	Code      string
	Name      string
	LastLat   float64
	LastLng   float64
	LocatedAt *int64
	Remark    string
	CreatedAt *int64
	UpdatedAt *int64
}

// IronMutateInput defines the create/update iron-cow input.
type IronMutateInput struct {
	Code   string
	Name   string
	Remark string
}

// ListIron returns one DB-side paged iron-cow page.
func (s *serviceImpl) ListIron(ctx context.Context, in *ListIronInput) (*ListIronOutput, error) {
	pageNum, pageSize := defaultPageNum, defaultPageSize
	model := dao.Iron.Ctx(ctx)
	if in != nil {
		pageNum, pageSize = normalizePagination(in.PageNum, in.PageSize)
		if keyword := strings.TrimSpace(in.Keyword); keyword != "" {
			like := "%" + keyword + "%"
			keywordFilter := model.Builder().
				WhereLike(dao.Iron.Columns().Code, like).
				WhereOrLike(dao.Iron.Columns().Name, like)
			model = model.Where(keywordFilter)
		}
	}

	total, err := model.Count()
	if err != nil {
		return nil, bizerr.WrapCode(err, CodeIronQueryFailed)
	}

	rows := make([]*entitymodel.Iron, 0)
	err = model.
		OrderDesc(dao.Iron.Columns().Id).
		Page(pageNum, pageSize).
		Scan(&rows)
	if err != nil {
		return nil, bizerr.WrapCode(err, CodeIronQueryFailed)
	}

	list := make([]*IronItem, 0, len(rows))
	for _, row := range rows {
		list = append(list, &IronItem{
			Id:        row.Id,
			Code:      row.Code,
			Name:      row.Name,
			LastLat:   row.LastLat,
			LastLng:   row.LastLng,
			LocatedAt: apitime.Milli(row.LocatedAt),
			Remark:    row.Remark,
			CreatedAt: apitime.Milli(row.CreatedAt),
			UpdatedAt: apitime.Milli(row.UpdatedAt),
		})
	}
	return &ListIronOutput{List: list, Total: total}, nil
}

// CreateIron inserts one iron-cow after validation. The real-time location
// columns are intentionally not written here.
func (s *serviceImpl) CreateIron(ctx context.Context, in *IronMutateInput) (int64, error) {
	code, name, err := validateIron(in)
	if err != nil {
		return 0, err
	}

	taken, err := s.ironCodeTaken(ctx, code, 0)
	if err != nil {
		return 0, err
	}
	if taken {
		return 0, bizerr.NewCode(CodeIronCodeExists)
	}

	id, err := dao.Iron.Ctx(ctx).Data(do.Iron{
		Code:   code,
		Name:   name,
		Remark: strings.TrimSpace(in.Remark),
	}).InsertAndGetId()
	if err != nil {
		return 0, bizerr.WrapCode(err, CodeIronWriteFailed)
	}
	return id, nil
}

// UpdateIron modifies one iron-cow's code, name and remark after validation. The
// real-time location columns are left untouched.
func (s *serviceImpl) UpdateIron(ctx context.Context, id int64, in *IronMutateInput) error {
	if id <= 0 {
		return bizerr.NewCode(CodeIronIDRequired)
	}
	code, name, err := validateIron(in)
	if err != nil {
		return err
	}

	exists, err := s.ironExists(ctx, id)
	if err != nil {
		return err
	}
	if !exists {
		return bizerr.NewCode(CodeIronNotFound)
	}

	taken, err := s.ironCodeTaken(ctx, code, id)
	if err != nil {
		return err
	}
	if taken {
		return bizerr.NewCode(CodeIronCodeExists)
	}

	_, err = dao.Iron.Ctx(ctx).Where(do.Iron{Id: id}).Data(do.Iron{
		Code:   code,
		Name:   name,
		Remark: strings.TrimSpace(in.Remark),
	}).Update()
	if err != nil {
		return bizerr.WrapCode(err, CodeIronWriteFailed)
	}
	return nil
}

// DeleteIron soft-deletes one iron-cow registration.
func (s *serviceImpl) DeleteIron(ctx context.Context, id int64) error {
	if id <= 0 {
		return bizerr.NewCode(CodeIronIDRequired)
	}
	exists, err := s.ironExists(ctx, id)
	if err != nil {
		return err
	}
	if !exists {
		return bizerr.NewCode(CodeIronNotFound)
	}

	_, err = dao.Iron.Ctx(ctx).Where(do.Iron{Id: id}).Delete()
	if err != nil {
		return bizerr.WrapCode(err, CodeIronWriteFailed)
	}
	return nil
}

// ironExists reports whether an active iron-cow with id exists.
func (s *serviceImpl) ironExists(ctx context.Context, id int64) (bool, error) {
	count, err := dao.Iron.Ctx(ctx).Where(do.Iron{Id: id}).Count()
	if err != nil {
		return false, bizerr.WrapCode(err, CodeIronQueryFailed)
	}
	return count > 0, nil
}

// ironCodeTaken reports whether code is owned by another active iron-cow. When
// excludeID is positive that iron-cow is excluded so an unchanged code on update
// does not collide with itself.
func (s *serviceImpl) ironCodeTaken(ctx context.Context, code string, excludeID int64) (bool, error) {
	model := dao.Iron.Ctx(ctx).Where(do.Iron{Code: code})
	if excludeID > 0 {
		model = model.WhereNot(dao.Iron.Columns().Id, excludeID)
	}
	count, err := model.Count()
	if err != nil {
		return false, bizerr.WrapCode(err, CodeIronQueryFailed)
	}
	return count > 0, nil
}

// validateIron trims and validates the required iron-cow code and name.
func validateIron(in *IronMutateInput) (code string, name string, err error) {
	if in == nil {
		return "", "", bizerr.NewCode(CodeIronCodeRequired)
	}
	code = strings.TrimSpace(in.Code)
	if code == "" {
		return "", "", bizerr.NewCode(CodeIronCodeRequired)
	}
	name = strings.TrimSpace(in.Name)
	if name == "" {
		return "", "", bizerr.NewCode(CodeIronNameRequired)
	}
	return code, name, nil
}
