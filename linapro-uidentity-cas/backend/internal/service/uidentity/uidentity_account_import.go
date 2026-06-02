// This file implements legacy account workbook validation and import using
// plugin-owned account, account detail, unit, and container tables.

package uidentity

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/gogf/gf/v2/util/gconv"
	"github.com/xuri/excelize/v2"

	"lina-core/pkg/bizerr"
	"lina-plugin-linapro-uidentity-cas/backend/internal/dao"
	"lina-plugin-linapro-uidentity-cas/backend/internal/model/do"
	"lina-plugin-linapro-uidentity-cas/backend/internal/model/entity"
)

const (
	defaultAccountImportLimit = 1000
	accountImportSheet        = "Sheet1"
	accountImportSource       = "import"
)

type accountImportWorkbook struct {
	rows []map[string]string
}

// CheckAccountImport validates one legacy account import workbook.
func (s *serviceImpl) CheckAccountImport(ctx context.Context, in AccountImportInput) (*AccountImportCheckOutput, error) {
	workbook, err := s.readAccountImportWorkbook(ctx, in)
	if err != nil {
		return nil, err
	}
	return &AccountImportCheckOutput{Rows: len(workbook.rows)}, nil
}

// ImportAccounts imports or updates accounts from one legacy workbook.
func (s *serviceImpl) ImportAccounts(ctx context.Context, in AccountImportInput) (*AccountImportOutput, error) {
	workbook, err := s.readAccountImportWorkbook(ctx, in)
	if err != nil {
		return nil, err
	}
	unitIDs, err := s.importUnitIDMap(ctx)
	if err != nil {
		return nil, err
	}
	containerIDs, err := s.importContainerIDMap(ctx)
	if err != nil {
		return nil, err
	}
	output := &AccountImportOutput{}
	for _, row := range workbook.rows {
		if err := s.importAccountRow(ctx, row, unitIDs, containerIDs); err != nil {
			output.FailedNumber = append(output.FailedNumber, row["number"])
			continue
		}
		output.Success++
	}
	return output, nil
}

func (s *serviceImpl) readAccountImportWorkbook(ctx context.Context, in AccountImportInput) (*accountImportWorkbook, error) {
	filepath := strings.TrimSpace(in.Filepath)
	if filepath == "" {
		return nil, bizerr.NewCode(CodeImportInvalid)
	}
	limit := in.Limit
	if limit <= 0 {
		configLimit, err := s.configSvc.Int(ctx, "import.accountLimit", defaultAccountImportLimit)
		if err != nil {
			return nil, err
		}
		limit = configLimit
	}
	if limit <= 0 {
		limit = defaultAccountImportLimit
	}
	file, err := excelize.OpenFile(filepath)
	if err != nil {
		return nil, bizerr.WrapCode(err, CodeImportInvalid)
	}
	defer func() {
		_ = file.Close()
	}()
	rawRows, err := file.GetRows(accountImportSheet)
	if err != nil {
		return nil, bizerr.WrapCode(err, CodeImportInvalid)
	}
	if len(rawRows) == 0 {
		return nil, bizerr.NewCode(CodeImportInvalid)
	}
	headers := normalizeImportHeaders(rawRows[0])
	if !hasImportHeader(headers, "number") {
		return nil, bizerr.NewCode(CodeImportInvalid)
	}
	rows := make([]map[string]string, 0, len(rawRows)-1)
	for index, rawRow := range rawRows[1:] {
		if importRowBlank(rawRow) {
			continue
		}
		row := make(map[string]string, len(headers))
		for i, header := range headers {
			if header == "" || i >= len(rawRow) {
				continue
			}
			row[header] = strings.TrimSpace(rawRow[i])
		}
		if strings.TrimSpace(row["number"]) == "" {
			return nil, bizerr.WrapCode(fmt.Errorf("row %d number is empty", index+2), CodeImportInvalid)
		}
		rows = append(rows, row)
		if len(rows) > limit {
			return nil, bizerr.WrapCode(fmt.Errorf("account import rows exceed limit %d", limit), CodeImportInvalid)
		}
	}
	return &accountImportWorkbook{rows: rows}, nil
}

func normalizeImportHeaders(raw []string) []string {
	headers := make([]string, 0, len(raw))
	for _, header := range raw {
		headers = append(headers, strings.TrimSpace(header))
	}
	return headers
}

func hasImportHeader(headers []string, expected string) bool {
	for _, header := range headers {
		if header == expected {
			return true
		}
	}
	return false
}

func importRowBlank(row []string) bool {
	if len(row) <= 1 {
		return true
	}
	for _, value := range row {
		if strings.TrimSpace(value) != "" {
			return false
		}
	}
	return true
}

func (s *serviceImpl) importAccountRow(ctx context.Context, row map[string]string, unitIDs map[string]int64, containerIDs map[string]int64) error {
	number := strings.TrimSpace(row["number"])
	if number == "" {
		return bizerr.NewCode(CodeImportInvalid)
	}
	account, err := s.accountByNumberAllowMissing(ctx, number)
	if err != nil {
		return err
	}
	if account == nil {
		return s.createImportedAccount(ctx, row, unitIDs, containerIDs)
	}
	return s.updateImportedAccount(ctx, account, row, unitIDs, containerIDs)
}

func (s *serviceImpl) createImportedAccount(ctx context.Context, row map[string]string, unitIDs map[string]int64, containerIDs map[string]int64) error {
	tenantID, actorID := s.baseOwnedDO(ctx, true)
	phone := strings.TrimSpace(row["phone"])
	if phone == "" {
		phone = strings.TrimSpace(row["number"])
	}
	accountData := do.Account{
		TenantId:    tenantID,
		Number:      strings.TrimSpace(row["number"]),
		Name:        strings.TrimSpace(row["name"]),
		Phone:       phone,
		EffectAt:    importEffectAt(row["effect_at"]),
		ExpireAt:    importExpireAt(row["expire_at"]),
		ContainerId: containerIDs[strings.TrimSpace(row["container_name"])],
		UnitId:      unitIDs[strings.TrimSpace(row["unit_code"])],
		Status:      importInt(row["status"]),
		CreatedBy:   actorID,
		UpdatedBy:   actorID,
	}
	accountID, err := s.createAccountWithAudit(ctx, accountData)
	if err != nil {
		return err
	}
	detailData := importedAccountDetailDO(ctx, s, accountID, row, true)
	return s.createAccountDetailWithAudit(ctx, detailData, accountID)
}

func (s *serviceImpl) updateImportedAccount(ctx context.Context, account *entity.Account, row map[string]string, unitIDs map[string]int64, containerIDs map[string]int64) error {
	accountData := do.Account{UpdatedBy: s.actorID(ctx)}
	if value := strings.TrimSpace(row["name"]); value != "" && value != account.Name {
		accountData.Name = value
	}
	if value := strings.TrimSpace(row["phone"]); value != "" && value != account.Phone {
		accountData.Phone = value
	}
	if value := strings.TrimSpace(row["status"]); value != "" {
		accountData.Status = importInt(value)
	}
	if value := strings.TrimSpace(row["effect_at"]); value != "" {
		accountData.EffectAt = importEffectAt(value)
	}
	if value := strings.TrimSpace(row["expire_at"]); value != "" {
		accountData.ExpireAt = importExpireAt(value)
	}
	if value := strings.TrimSpace(row["container_name"]); value != "" {
		accountData.ContainerId = containerIDs[value]
	}
	if value := strings.TrimSpace(row["unit_code"]); value != "" {
		accountData.UnitId = unitIDs[value]
	}
	if err := s.updateAccountWithAudit(ctx, account.Id, accountData); err != nil {
		return err
	}
	count, err := s.tenantFilter.Apply(ctx, dao.AccountDetail.Ctx(ctx), "").
		Where(dao.AccountDetail.Columns().AccountId, account.Id).
		Count()
	if err != nil {
		return err
	}
	detailData := importedAccountDetailDO(ctx, s, account.Id, row, count == 0)
	if count == 0 {
		return s.createAccountDetailWithAudit(ctx, detailData, account.Id)
	}
	return s.updateAccountDetailWithAudit(ctx, account.Id, detailData)
}

func importedAccountDetailDO(ctx context.Context, s *serviceImpl, accountID int64, row map[string]string, create bool) do.AccountDetail {
	tenantID, actorID := s.baseOwnedDO(ctx, create)
	data := do.AccountDetail{UpdatedBy: actorID}
	if create {
		data.TenantId = tenantID
		data.AccountId = accountID
		data.CreatedBy = actorID
		data.Source = accountImportSource
	}
	if value := importBirthday(row["birthday"]); value != "" {
		data.Birthday = value
	}
	copyImportString(row, "email", &data.Email)
	copyImportString(row, "qq", &data.Qq)
	copyImportString(row, "wechat", &data.Wechat)
	copyImportString(row, "idcard", &data.Idcard)
	copyImportString(row, "avatar", &data.Avatar)
	copyImportString(row, "grade", &data.Grade)
	copyImportString(row, "college", &data.College)
	copyImportString(row, "college_code", &data.CollegeCode)
	copyImportString(row, "campus", &data.Campus)
	copyImportString(row, "school_system", &data.SchoolSystem)
	copyImportString(row, "graduated_at", &data.GraduatedAt)
	copyImportString(row, "major", &data.Major)
	copyImportString(row, "class_name", &data.ClassName)
	copyImportString(row, "face", &data.Face)
	if value := strings.TrimSpace(row["gender"]); value != "" {
		data.Gender = importInt(value)
	}
	return data
}

func copyImportString(row map[string]string, key string, target *any) {
	if value := strings.TrimSpace(row[key]); value != "" {
		*target = value
	}
}

func (s *serviceImpl) accountByNumberAllowMissing(ctx context.Context, number string) (*entity.Account, error) {
	var account *entity.Account
	err := s.tenantFilter.Apply(ctx, dao.Account.Ctx(ctx), "").
		Where(dao.Account.Columns().Number, number).
		Scan(&account)
	if err != nil {
		return nil, err
	}
	return account, nil
}

func (s *serviceImpl) importUnitIDMap(ctx context.Context) (map[string]int64, error) {
	result := make(map[string]int64)
	rows, err := s.tenantFilter.Apply(ctx, dao.Unit.Ctx(ctx), "").
		Fields(dao.Unit.Columns().Id, dao.Unit.Columns().Code).
		All()
	if err != nil {
		return nil, err
	}
	for _, row := range rows {
		result[row[dao.Unit.Columns().Code].String()] = row[dao.Unit.Columns().Id].Int64()
	}
	return result, nil
}

func (s *serviceImpl) importContainerIDMap(ctx context.Context) (map[string]int64, error) {
	result := make(map[string]int64)
	rows, err := s.tenantFilter.Apply(ctx, dao.Container.Ctx(ctx), "").
		Fields(dao.Container.Columns().Id, dao.Container.Columns().Name).
		All()
	if err != nil {
		return nil, err
	}
	for _, row := range rows {
		result[row[dao.Container.Columns().Name].String()] = row[dao.Container.Columns().Id].Int64()
	}
	return result, nil
}

func importInt(value string) int {
	parsed, err := strconv.Atoi(strings.TrimSpace(value))
	if err != nil {
		return 0
	}
	return parsed
}

func importEffectAt(value string) *time.Time {
	if strings.TrimSpace(value) == "" {
		now := time.Now()
		return &now
	}
	parsed := parseImportTime(value)
	return &parsed
}

func importExpireAt(value string) *time.Time {
	if strings.TrimSpace(value) == "" {
		expireAt := time.Date(2099, 12, 31, 23, 59, 59, 0, time.Local)
		return &expireAt
	}
	parsed := parseImportTime(value)
	return &parsed
}

func parseImportTime(value string) time.Time {
	trimmed := strings.TrimSpace(value)
	for _, layout := range []string{time.RFC3339, "2006-01-02 15:04:05", "2006-01-02"} {
		parsed, err := time.ParseInLocation(layout, trimmed, time.Local)
		if err == nil {
			return parsed
		}
	}
	if numeric := gconv.Float64(trimmed); numeric > 0 {
		converted, err := excelize.ExcelDateToTime(numeric, false)
		if err == nil {
			return converted
		}
	}
	return time.Now()
}

func importBirthday(value string) string {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return ""
	}
	parsed := parseImportTime(trimmed)
	if parsed.IsZero() {
		return trimmed
	}
	return parsed.Format("2006-01-02")
}
