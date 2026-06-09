// cattle_niu.go implements cattle listing, detail, creation, update and
// cascade-deletion. Listing runs DB-side filtering/sorting/pagination and then
// batch-assembles the linked college names and main-card binding flags in two
// bounded queries to avoid N+1. Mutations enforce code uniqueness, type/subtype
// and college-existence validation with bounded queries. Deletion soft-deletes
// the cattle and its unique main card in one transaction.

package cattle

import (
	"context"
	"strings"
	"time"

	"github.com/gogf/gf/v2/database/gdb"

	"lina-core/pkg/apitime"
	"lina-core/pkg/bizerr"
	"lina-plugin-sicau-niu/backend/internal/dao"
	"lina-plugin-sicau-niu/backend/internal/model/do"
	entitymodel "lina-plugin-sicau-niu/backend/internal/model/entity"
)

// ListNiuInput defines the operator cattle list query.
type ListNiuInput struct {
	// Keyword is the optional fuzzy match applied to the cattle code or name.
	Keyword string
	// NiuType optionally filters cattle by type; empty lists all.
	NiuType string
	// PageNum is the requested page number; defaults to 1 when non-positive.
	PageNum int
	// PageSize is the requested page size; defaults to 10 and is capped at 100.
	PageSize int
}

// ListNiuOutput defines the operator cattle list result.
type ListNiuOutput struct {
	// List holds the current page of cattle.
	List []*NiuItem
	// Total is the total matched cattle count.
	Total int
}

// NiuItem defines one cattle row projected for the operator console, including
// the batch-assembled college name and card-binding flag.
type NiuItem struct {
	Id              int64
	Code            string
	NiuType         string
	SpecialSubtype  string
	Name            string
	CollegeId       int64
	CollegeName     string
	Lat             float64
	Lng             float64
	OnlineAt        *int64
	VisibleWeekdays string
	VisibleStart    string
	VisibleEnd      string
	Status          string
	HasCard         bool
	CreatedAt       *int64
	UpdatedAt       *int64
}

// NiuMutateInput defines the create/update cattle input. OnlineAt is a Unix
// timestamp in milliseconds, nil when unset.
type NiuMutateInput struct {
	Code            string
	NiuType         string
	SpecialSubtype  string
	Name            string
	CollegeId       int64
	Lat             float64
	Lng             float64
	OnlineAt        *int64
	VisibleWeekdays string
	VisibleStart    string
	VisibleEnd      string
}

// ListNiu returns one DB-side paged cattle page with batch-assembled relations.
func (s *serviceImpl) ListNiu(ctx context.Context, in *ListNiuInput) (*ListNiuOutput, error) {
	pageNum, pageSize := defaultPageNum, defaultPageSize
	model := dao.Niu.Ctx(ctx)
	if in != nil {
		pageNum, pageSize = normalizePagination(in.PageNum, in.PageSize)
		if keyword := strings.TrimSpace(in.Keyword); keyword != "" {
			like := "%" + keyword + "%"
			keywordFilter := model.Builder().
				WhereLike(dao.Niu.Columns().Code, like).
				WhereOrLike(dao.Niu.Columns().Name, like)
			model = model.Where(keywordFilter)
		}
		if niuType := strings.TrimSpace(in.NiuType); niuType != "" {
			model = model.Where(do.Niu{NiuType: niuType})
		}
	}

	total, err := model.Count()
	if err != nil {
		return nil, bizerr.WrapCode(err, CodeNiuQueryFailed)
	}

	rows := make([]*entitymodel.Niu, 0)
	err = model.
		OrderDesc(dao.Niu.Columns().Id).
		Page(pageNum, pageSize).
		Scan(&rows)
	if err != nil {
		return nil, bizerr.WrapCode(err, CodeNiuQueryFailed)
	}

	list, err := s.assembleNiuItems(ctx, rows)
	if err != nil {
		return nil, err
	}
	return &ListNiuOutput{List: list, Total: total}, nil
}

// GetNiu returns one cattle detail with its batch-assembled relations.
func (s *serviceImpl) GetNiu(ctx context.Context, id int64) (*NiuItem, error) {
	if id <= 0 {
		return nil, bizerr.NewCode(CodeNiuIDRequired)
	}
	var row *entitymodel.Niu
	err := dao.Niu.Ctx(ctx).Where(do.Niu{Id: id}).Scan(&row)
	if err != nil {
		return nil, bizerr.WrapCode(err, CodeNiuQueryFailed)
	}
	if row == nil {
		return nil, bizerr.NewCode(CodeNiuNotFound)
	}
	items, err := s.assembleNiuItems(ctx, []*entitymodel.Niu{row})
	if err != nil {
		return nil, err
	}
	return items[0], nil
}

// CreateNiu inserts one cattle after validation. Status defaults to inactive.
func (s *serviceImpl) CreateNiu(ctx context.Context, in *NiuMutateInput) (int64, error) {
	data, err := s.validateNiu(ctx, in, 0)
	if err != nil {
		return 0, err
	}
	data.Status = NiuStatusInactive.String()

	id, err := dao.Niu.Ctx(ctx).Data(data).InsertAndGetId()
	if err != nil {
		return 0, bizerr.WrapCode(err, CodeNiuWriteFailed)
	}
	return id, nil
}

// UpdateNiu modifies one cattle after existence and validation. Status is left
// untouched so the activation flow (C3) owns status transitions.
func (s *serviceImpl) UpdateNiu(ctx context.Context, id int64, in *NiuMutateInput) error {
	if id <= 0 {
		return bizerr.NewCode(CodeNiuIDRequired)
	}
	exists, err := s.NiuExists(ctx, id)
	if err != nil {
		return err
	}
	if !exists {
		return bizerr.NewCode(CodeNiuNotFound)
	}

	data, err := s.validateNiu(ctx, in, id)
	if err != nil {
		return err
	}

	_, err = dao.Niu.Ctx(ctx).Where(do.Niu{Id: id}).Data(data).Update()
	if err != nil {
		return bizerr.WrapCode(err, CodeNiuWriteFailed)
	}
	return nil
}

// DeleteNiu soft-deletes one cattle and cascade soft-deletes its main card in one
// transaction so no dangling card remains.
func (s *serviceImpl) DeleteNiu(ctx context.Context, id int64) error {
	if id <= 0 {
		return bizerr.NewCode(CodeNiuIDRequired)
	}
	exists, err := s.NiuExists(ctx, id)
	if err != nil {
		return err
	}
	if !exists {
		return bizerr.NewCode(CodeNiuNotFound)
	}

	err = dao.Niu.Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {
		if _, txErr := dao.Niu.Ctx(ctx).Where(do.Niu{Id: id}).Delete(); txErr != nil {
			return bizerr.WrapCode(txErr, CodeNiuWriteFailed)
		}
		if _, txErr := dao.Card.Ctx(ctx).Where(do.Card{NiuId: id}).Delete(); txErr != nil {
			return bizerr.WrapCode(txErr, CodeNiuWriteFailed)
		}
		return nil
	})
	if err != nil {
		return err
	}
	return nil
}

// NiuExists reports whether an active cattle with id exists.
func (s *serviceImpl) NiuExists(ctx context.Context, id int64) (bool, error) {
	if id <= 0 {
		return false, nil
	}
	count, err := dao.Niu.Ctx(ctx).Where(do.Niu{Id: id}).Count()
	if err != nil {
		return false, bizerr.WrapCode(err, CodeNiuQueryFailed)
	}
	return count > 0, nil
}

// assembleNiuItems projects cattle rows to operator items, batch-loading the
// linked college names and main-card binding flags in two bounded queries keyed
// by the current page's college and cattle IDs to avoid N+1.
func (s *serviceImpl) assembleNiuItems(ctx context.Context, rows []*entitymodel.Niu) ([]*NiuItem, error) {
	collegeNames, err := s.batchCollegeNames(ctx, rows)
	if err != nil {
		return nil, err
	}
	cardBound, err := s.batchCardBound(ctx, rows)
	if err != nil {
		return nil, err
	}

	list := make([]*NiuItem, 0, len(rows))
	for _, row := range rows {
		list = append(list, &NiuItem{
			Id:              row.Id,
			Code:            row.Code,
			NiuType:         row.NiuType,
			SpecialSubtype:  row.SpecialSubtype,
			Name:            row.Name,
			CollegeId:       row.CollegeId,
			CollegeName:     collegeNames[row.CollegeId],
			Lat:             row.Lat,
			Lng:             row.Lng,
			OnlineAt:        apitime.Milli(row.OnlineAt),
			VisibleWeekdays: row.VisibleWeekdays,
			VisibleStart:    row.VisibleStart,
			VisibleEnd:      row.VisibleEnd,
			Status:          row.Status,
			HasCard:         cardBound[row.Id],
			CreatedAt:       apitime.Milli(row.CreatedAt),
			UpdatedAt:       apitime.Milli(row.UpdatedAt),
		})
	}
	return list, nil
}

// batchCollegeNames returns a college-ID to name map for the distinct positive
// college IDs on the page, fetched in one projected query. An empty key set
// short-circuits without a query.
func (s *serviceImpl) batchCollegeNames(ctx context.Context, rows []*entitymodel.Niu) (map[int64]string, error) {
	ids := distinctPositive(rows, func(row *entitymodel.Niu) int64 { return row.CollegeId })
	names := make(map[int64]string, len(ids))
	if len(ids) == 0 {
		return names, nil
	}

	collegeRows := make([]*entitymodel.College, 0, len(ids))
	err := dao.College.Ctx(ctx).
		Fields(dao.College.Columns().Id, dao.College.Columns().Name).
		WhereIn(dao.College.Columns().Id, ids).
		Scan(&collegeRows)
	if err != nil {
		return nil, bizerr.WrapCode(err, CodeNiuQueryFailed)
	}
	for _, row := range collegeRows {
		names[row.Id] = row.Name
	}
	return names, nil
}

// batchCardBound returns a cattle-ID to bound flag map for the cattle on the page
// that own an active card, fetched in one projected query over card.niu_id. An
// empty page short-circuits without a query.
func (s *serviceImpl) batchCardBound(ctx context.Context, rows []*entitymodel.Niu) (map[int64]bool, error) {
	ids := distinctPositive(rows, func(row *entitymodel.Niu) int64 { return row.Id })
	bound := make(map[int64]bool, len(ids))
	if len(ids) == 0 {
		return bound, nil
	}

	cardRows := make([]*entitymodel.Card, 0, len(ids))
	err := dao.Card.Ctx(ctx).
		Fields(dao.Card.Columns().NiuId).
		WhereIn(dao.Card.Columns().NiuId, ids).
		Scan(&cardRows)
	if err != nil {
		return nil, bizerr.WrapCode(err, CodeNiuQueryFailed)
	}
	for _, row := range cardRows {
		bound[row.NiuId] = true
	}
	return bound, nil
}

// distinctPositive collects the distinct positive keys extracted from rows by
// key, preserving first-seen order, for use as a bounded WHERE IN key set.
func distinctPositive(rows []*entitymodel.Niu, key func(*entitymodel.Niu) int64) []int64 {
	seen := make(map[int64]struct{}, len(rows))
	ids := make([]int64, 0, len(rows))
	for _, row := range rows {
		id := key(row)
		if id <= 0 {
			continue
		}
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		ids = append(ids, id)
	}
	return ids
}

// validateNiu validates the mutate input and builds the persisted DO. It enforces
// code presence/uniqueness, type/subtype rules and college-existence
// rules. excludeID is the cattle excluded from the code-uniqueness check on
// update so an unchanged code does not collide with itself.
func (s *serviceImpl) validateNiu(ctx context.Context, in *NiuMutateInput, excludeID int64) (do.Niu, error) {
	if in == nil {
		return do.Niu{}, bizerr.NewCode(CodeNiuCodeRequired)
	}
	code := strings.TrimSpace(in.Code)
	if code == "" {
		return do.Niu{}, bizerr.NewCode(CodeNiuCodeRequired)
	}

	niuType := NiuType(strings.TrimSpace(in.NiuType))
	if !niuType.valid() {
		return do.Niu{}, bizerr.NewCode(CodeNiuTypeInvalid)
	}

	subtype, name, collegeID, err := s.validateNiuTypeShape(ctx, niuType, in)
	if err != nil {
		return do.Niu{}, err
	}

	taken, err := s.niuCodeTaken(ctx, code, excludeID)
	if err != nil {
		return do.Niu{}, err
	}
	if taken {
		return do.Niu{}, bizerr.NewCode(CodeNiuCodeExists)
	}

	return do.Niu{
		Code:            code,
		NiuType:         niuType.String(),
		SpecialSubtype:  subtype,
		Name:            name,
		CollegeId:       collegeID,
		Lat:             in.Lat,
		Lng:             in.Lng,
		OnlineAt:        milliToTime(in.OnlineAt),
		VisibleWeekdays: strings.TrimSpace(in.VisibleWeekdays),
		VisibleStart:    strings.TrimSpace(in.VisibleStart),
		VisibleEnd:      strings.TrimSpace(in.VisibleEnd),
	}, nil
}

// validateNiuTypeShape validates the subtype/name/college shape implied by the
// cattle type and returns the resolved subtype, name and college ID to persist.
// Common cattle clear subtype/college; special cattle require a name and a valid
// subtype; college subtype cattle require an existing linked college.
func (s *serviceImpl) validateNiuTypeShape(
	ctx context.Context,
	niuType NiuType,
	in *NiuMutateInput,
) (subtype string, name string, collegeID int64, err error) {
	name = strings.TrimSpace(in.Name)
	if !niuType.requiresSubtype() {
		// Common cattle: drop subtype and college; name stays optional.
		return "", name, 0, nil
	}

	if name == "" {
		return "", "", 0, bizerr.NewCode(CodeNiuNameRequired)
	}
	sub := SpecialSubtype(strings.TrimSpace(in.SpecialSubtype))
	if !sub.valid() {
		return "", "", 0, bizerr.NewCode(CodeNiuSubtypeInvalid)
	}
	if !sub.requiresCollege() {
		return sub.String(), name, 0, nil
	}

	if in.CollegeId <= 0 {
		return "", "", 0, bizerr.NewCode(CodeNiuCollegeRequired)
	}
	collegeExists, err := s.collegeSvc.Exists(ctx, in.CollegeId)
	if err != nil {
		return "", "", 0, err
	}
	if !collegeExists {
		return "", "", 0, bizerr.NewCode(CodeNiuCollegeInvalid)
	}
	return sub.String(), name, in.CollegeId, nil
}

// niuCodeTaken reports whether code is owned by another active cattle. When
// excludeID is positive that cattle is excluded so an unchanged code on update
// does not collide with itself.
func (s *serviceImpl) niuCodeTaken(ctx context.Context, code string, excludeID int64) (bool, error) {
	model := dao.Niu.Ctx(ctx).Where(do.Niu{Code: code})
	if excludeID > 0 {
		model = model.WhereNot(dao.Niu.Columns().Id, excludeID)
	}
	count, err := model.Count()
	if err != nil {
		return false, bizerr.WrapCode(err, CodeNiuQueryFailed)
	}
	return count > 0, nil
}

// milliToTime converts an optional Unix-millisecond timestamp to a time pointer
// for storage. A nil or non-positive value is persisted as nil so an unset
// online time stays NULL.
func milliToTime(millis *int64) *time.Time {
	if millis == nil || *millis <= 0 {
		return nil
	}
	t := time.UnixMilli(*millis)
	return &t
}
