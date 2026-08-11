// cattle_niu_import.go implements bounded transactional cattle batch import.
package cattle

import (
	"context"
	"strconv"
	"strings"
	"time"

	"github.com/gogf/gf/v2/database/gdb"

	"lina-core/pkg/bizerr"
	"lina-plugin-sicau-niu/backend/internal/dao"
	"lina-plugin-sicau-niu/backend/internal/model/do"
	entitymodel "lina-plugin-sicau-niu/backend/internal/model/entity"
)

const maxNiuImportItems = 200

type ImportNiuInput struct {
	Items     []*NiuMutateInput
	Overwrite bool
}

type ImportNiuOutput struct {
	Created int
	Updated int
	Skipped int
}

// ImportNiu validates relations and existing codes with two bounded set queries.
func (s *serviceImpl) ImportNiu(ctx context.Context, in *ImportNiuInput) (*ImportNiuOutput, error) {
	if in == nil || len(in.Items) == 0 || len(in.Items) > maxNiuImportItems {
		return nil, bizerr.NewCode(CodeNiuImportInvalid)
	}
	codes := make([]string, 0, len(in.Items))
	collegeIDs := make([]int64, 0, len(in.Items))
	seenCodes := make(map[string]bool, len(in.Items))
	for _, item := range in.Items {
		if item == nil {
			return nil, bizerr.NewCode(CodeNiuImportInvalid)
		}
		code := strings.TrimSpace(item.Code)
		if code == "" || seenCodes[code] {
			return nil, bizerr.NewCode(CodeNiuCodeExists)
		}
		seenCodes[code] = true
		codes = append(codes, code)
		if item.CollegeId > 0 {
			collegeIDs = append(collegeIDs, item.CollegeId)
		}
	}
	collegeRows := make([]*entitymodel.College, 0)
	if len(collegeIDs) > 0 {
		if err := dao.College.Ctx(ctx).Fields(dao.College.Columns().Id).WhereIn(dao.College.Columns().Id, collegeIDs).Scan(&collegeRows); err != nil {
			return nil, bizerr.WrapCode(err, CodeNiuQueryFailed)
		}
	}
	colleges := make(map[int64]bool, len(collegeRows))
	for _, row := range collegeRows {
		colleges[row.Id] = true
	}
	existingRows := make([]*entitymodel.Niu, 0)
	if err := dao.Niu.Ctx(ctx).Fields(dao.Niu.Columns().Id, dao.Niu.Columns().Code).WhereIn(dao.Niu.Columns().Code, codes).Scan(&existingRows); err != nil {
		return nil, bizerr.WrapCode(err, CodeNiuQueryFailed)
	}
	existing := make(map[string]int64, len(existingRows))
	for _, row := range existingRows {
		existing[row.Code] = row.Id
	}
	data := make([]do.Niu, 0, len(in.Items))
	for _, item := range in.Items {
		row, err := validateImportNiu(item, colleges)
		if err != nil {
			return nil, err
		}
		data = append(data, row)
	}
	result := &ImportNiuOutput{}
	err := dao.Niu.Transaction(ctx, func(ctx context.Context, _ gdb.TX) error {
		for i, row := range data {
			code := codes[i]
			if id := existing[code]; id > 0 {
				if !in.Overwrite {
					result.Skipped++
					continue
				}
				if _, err := dao.Niu.Ctx(ctx).Where(do.Niu{Id: id}).Data(row).Update(); err != nil {
					return bizerr.WrapCode(err, CodeNiuWriteFailed)
				}
				result.Updated++
				continue
			}
			row.Status = NiuStatusInactive.String()
			if _, err := dao.Niu.Ctx(ctx).Data(row).Insert(); err != nil {
				return bizerr.WrapCode(err, CodeNiuWriteFailed)
			}
			result.Created++
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return result, nil
}

func validateImportNiu(in *NiuMutateInput, colleges map[int64]bool) (do.Niu, error) {
	niuType := NiuType(strings.TrimSpace(in.NiuType))
	if !niuType.valid() || in.Lat < -90 || in.Lat > 90 || in.Lng < -180 || in.Lng > 180 || (in.Lat == 0 && in.Lng == 0) || !validVisibility(in.VisibleWeekdays, in.VisibleStart, in.VisibleEnd) {
		return do.Niu{}, bizerr.NewCode(CodeNiuImportInvalid)
	}
	name := strings.TrimSpace(in.Name)
	subtype := ""
	collegeID := int64(0)
	if niuType.requiresSubtype() {
		sub := SpecialSubtype(strings.TrimSpace(in.SpecialSubtype))
		if name == "" {
			return do.Niu{}, bizerr.NewCode(CodeNiuNameRequired)
		}
		if !sub.valid() {
			return do.Niu{}, bizerr.NewCode(CodeNiuSubtypeInvalid)
		}
		subtype = sub.String()
		if sub.requiresCollege() {
			if in.CollegeId <= 0 || !colleges[in.CollegeId] {
				return do.Niu{}, bizerr.NewCode(CodeNiuCollegeInvalid)
			}
			collegeID = in.CollegeId
		}
	}
	return do.Niu{
		Code: strings.TrimSpace(in.Code), NiuType: niuType.String(), SpecialSubtype: subtype,
		Name: name, CollegeId: collegeID, Lat: in.Lat, Lng: in.Lng, OnlineAt: milliToTime(in.OnlineAt),
		VisibleWeekdays: strings.TrimSpace(in.VisibleWeekdays), VisibleStart: strings.TrimSpace(in.VisibleStart), VisibleEnd: strings.TrimSpace(in.VisibleEnd),
	}, nil
}

func validVisibility(weekdays, start, end string) bool {
	weekdays = strings.TrimSpace(weekdays)
	if weekdays != "" {
		for _, raw := range strings.Split(weekdays, ",") {
			day, err := strconv.Atoi(strings.TrimSpace(raw))
			if err != nil || day < 1 || day > 7 {
				return false
			}
		}
	}
	start, end = strings.TrimSpace(start), strings.TrimSpace(end)
	if (start == "") != (end == "") {
		return false
	}
	if start == "" {
		return true
	}
	_, startErr := time.Parse("15:04", start)
	_, endErr := time.Parse("15:04", end)
	return startErr == nil && endErr == nil && start != end
}
