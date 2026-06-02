// This file implements shared MySQL-side account upsert logic for Oracle-backed
// UIdentity scheduled jobs, keeping reads batched by page and audit writes
// plugin-local.

package jobs

import (
	"context"
	"strconv"
	"strings"
	"time"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"

	"lina-core/pkg/logger"
	"lina-plugin-linapro-uidentity-cas/backend/internal/dao"
	"lina-plugin-linapro-uidentity-cas/backend/internal/model/do"
	"lina-plugin-linapro-uidentity-cas/backend/internal/model/entity"
)

type accountSyncInput struct {
	number                 string
	name                   string
	phone                  string
	statusRaw              string
	unitCode               string
	containerName          string
	groupName              string
	requireNumericUnitCode bool
	detail                 accountDetailSyncInput
}

type accountDetailSyncInput struct {
	idcard       string
	birthday     string
	email        string
	gender       int
	qq           string
	wechat       string
	avatar       string
	source       string
	grade        string
	college      string
	collegeCode  string
	campus       string
	schoolSystem string
	graduatedAt  string
	major        string
	className    string
	face         string
}

type syncContext struct {
	tenantID      int
	containerID   int64
	groupID       int64
	unitsByCode   map[string]*entity.Unit
	newUserCutoff time.Time
}

func (s *serviceImpl) syncOracleAccounts(ctx context.Context, syncCtx *syncContext, inputs []*accountSyncInput) (jobRunStats, error) {
	validInputs, stats := prepareAccountSyncInputs(inputs)
	if len(validInputs) == 0 {
		return stats, nil
	}
	numbers := make(map[string]struct{}, len(validInputs))
	phones := make(map[string]struct{}, len(validInputs))
	for _, input := range validInputs {
		numbers[input.number] = struct{}{}
		if input.phone != "" {
			phones[input.phone] = struct{}{}
		}
	}
	accountsByNumber, err := legacyAccountsByNumbers(ctx, syncCtx.tenantID, stringsFromSet(numbers))
	if err != nil {
		return jobRunStats{}, err
	}
	accountIDs := make(map[int64]struct{}, len(accountsByNumber))
	for _, account := range accountsByNumber {
		if account != nil {
			accountIDs[account.Id] = struct{}{}
		}
	}
	detailsByAccountID, err := legacyAccountDetailsByIDs(ctx, syncCtx.tenantID, int64sFromSet(accountIDs))
	if err != nil {
		return jobRunStats{}, err
	}
	phoneOwners, err := legacyPhoneOwnersByPhone(ctx, syncCtx.tenantID, stringsFromSet(phones))
	if err != nil {
		return jobRunStats{}, err
	}
	for _, input := range validInputs {
		account := accountsByNumber[input.number]
		if account == nil {
			err = s.createOracleAccount(ctx, syncCtx, input, &stats)
		} else {
			err = s.updateOracleAccount(ctx, syncCtx, input, account, detailsByAccountID[account.Id], phoneOwners, &stats)
		}
		if err != nil {
			stats.errNum++
			logger.Warningf(ctx, "uidentity oracle account sync failed number=%s err=%v", input.number, err)
		}
	}
	return stats, nil
}

func (s *serviceImpl) accountSyncContext(ctx context.Context, tenantID int, containerName string, groupName string) (*syncContext, error) {
	containerID, err := legacyContainerIDByName(ctx, tenantID, containerName)
	if err != nil {
		return nil, err
	}
	groupID, err := legacyGroupIDByName(ctx, tenantID, groupName)
	if err != nil {
		return nil, err
	}
	units, err := legacyUnitsByCode(ctx, tenantID)
	if err != nil {
		return nil, err
	}
	cutoff, err := s.newUserCutoff(ctx)
	if err != nil {
		return nil, err
	}
	return &syncContext{
		tenantID:      tenantID,
		containerID:   containerID,
		groupID:       groupID,
		unitsByCode:   units,
		newUserCutoff: cutoff,
	}, nil
}

func (s *serviceImpl) createOracleAccount(ctx context.Context, syncCtx *syncContext, input *accountSyncInput, stats *jobRunStats) error {
	now := time.Now()
	phone := input.phone
	if phone == "" {
		phone = input.number
	}
	data := do.Account{
		TenantId:     syncCtx.tenantID,
		Number:       input.number,
		Name:         input.name,
		Phone:        phone,
		PasswordHash: hashAccountPassword(input.number),
		EffectAt:     &now,
		ExpireAt:     defaultExpireAt(),
		PassLevel:    0,
		ContainerId:  syncCtx.containerID,
		UnitId:       unitID(syncCtx.unitsByCode, input.unitCode),
		Status:       statusValue(input.statusRaw, true, syncCtx.newUserCutoff),
		CreatedBy:    int64(0),
		UpdatedBy:    int64(0),
	}
	accountID, err := dao.Account.Ctx(ctx).Data(data).InsertAndGetId()
	if err != nil && phone != input.number {
		data.Phone = input.number
		accountID, err = dao.Account.Ctx(ctx).Data(data).InsertAndGetId()
	}
	if err != nil {
		return err
	}
	if err := s.insertAccountAudit(ctx, syncCtx.tenantID, accountID, legacyAuditTableAccount, legacyAuditCreate, nil, data); err != nil {
		return err
	}
	detail := detailDO(syncCtx.tenantID, accountID, input.detail, true)
	if _, err = dao.AccountDetail.Ctx(ctx).Data(detail).Insert(); err != nil {
		return err
	}
	if err := s.insertAccountAudit(ctx, syncCtx.tenantID, accountID, legacyAuditTableDetail, legacyAuditCreate, nil, detail); err != nil {
		return err
	}
	if syncCtx.groupID > 0 {
		_, err = dao.AccountGroup.Ctx(ctx).Data(do.AccountGroup{
			TenantId:  syncCtx.tenantID,
			AccountId: accountID,
			GroupId:   syncCtx.groupID,
			CreatedBy: int64(0),
		}).Insert()
		if err != nil {
			return err
		}
	}
	stats.createNum++
	return nil
}

func (s *serviceImpl) updateOracleAccount(ctx context.Context, syncCtx *syncContext, input *accountSyncInput, account *entity.Account, detail *entity.AccountDetail, phoneOwners map[string]map[int64]struct{}, stats *jobRunStats) error {
	accountData := do.Account{UpdatedBy: int64(0)}
	accountChanged := false
	if input.name != "" && input.name != account.Name {
		accountData.Name = input.name
		accountChanged = true
	}
	if input.phone != "" && input.phone != input.number && input.phone != account.Phone && account.Status != 2 && account.Phone == account.Number {
		if !phoneUsedByOther(phoneOwners, input.phone, account.Id) {
			accountData.Phone = input.phone
			accountChanged = true
		}
	}
	if nextUnitID := unitID(syncCtx.unitsByCode, input.unitCode); nextUnitID > 0 && nextUnitID != account.UnitId {
		accountData.UnitId = nextUnitID
		accountChanged = true
	}
	if nextStatus := statusValue(input.statusRaw, false, syncCtx.newUserCutoff); nextStatus == 2 && nextStatus != account.Status {
		accountData.Status = nextStatus
		accountChanged = true
	}
	if syncCtx.containerID > 0 && syncCtx.containerID != account.ContainerId {
		accountData.ContainerId = syncCtx.containerID
		accountChanged = true
	}
	if accountChanged {
		if err := s.updateAccountFromJob(ctx, syncCtx, account, accountData); err != nil {
			return err
		}
		stats.updateAccountCount++
	}
	detailChanged, err := s.upsertDetailFromJob(ctx, syncCtx, account.Id, detail, input.detail, stats)
	if err != nil {
		return err
	}
	if accountChanged || detailChanged {
		stats.updateNum++
	}
	return nil
}

func (s *serviceImpl) updateAccountFromJob(ctx context.Context, syncCtx *syncContext, account *entity.Account, data do.Account) error {
	_, err := dao.Account.Ctx(ctx).
		Unscoped().
		Where(dao.Account.Columns().TenantId, syncCtx.tenantID).
		Where(dao.Account.Columns().Id, account.Id).
		OmitNilData().
		Data(data).
		Update()
	if err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "phone") {
			return nil
		}
		return err
	}
	return s.insertAccountAudit(ctx, syncCtx.tenantID, account.Id, legacyAuditTableAccount, legacyAuditUpdate, accountAuditRecord(account), mergeAccountAudit(account, data))
}

func (s *serviceImpl) upsertDetailFromJob(ctx context.Context, syncCtx *syncContext, accountID int64, detail *entity.AccountDetail, input accountDetailSyncInput, stats *jobRunStats) (bool, error) {
	if detail == nil {
		return false, nil
	}
	data, changed := changedDetailDO(detail, input)
	if !changed {
		return false, nil
	}
	_, err := dao.AccountDetail.Ctx(ctx).
		Where(dao.AccountDetail.Columns().TenantId, syncCtx.tenantID).
		Where(dao.AccountDetail.Columns().AccountId, accountID).
		OmitNilData().
		Data(data).
		Update()
	if err != nil {
		return false, err
	}
	stats.updateAccountDetailCount++
	return true, s.insertAccountAudit(ctx, syncCtx.tenantID, accountID, legacyAuditTableDetail, legacyAuditUpdate, detailAuditRecord(detail), mergeDetailAudit(detail, data))
}

func legacyGroupIDByName(ctx context.Context, tenantID int, name string) (int64, error) {
	var group *entity.Group
	err := dao.Group.Ctx(ctx).
		Fields(dao.Group.Columns().Id).
		Where(dao.Group.Columns().TenantId, tenantID).
		Where(dao.Group.Columns().Name, strings.TrimSpace(name)).
		Scan(&group)
	if err != nil {
		return 0, err
	}
	if group == nil || group.Id <= 0 {
		return 0, pluginJobUnsupported()
	}
	return group.Id, nil
}

func legacyAccountsByNumbers(ctx context.Context, tenantID int, numbers []string) (map[string]*entity.Account, error) {
	result := make(map[string]*entity.Account, len(numbers))
	if len(numbers) == 0 {
		return result, nil
	}
	var accounts []*entity.Account
	err := dao.Account.Ctx(ctx).
		Unscoped().
		Where(dao.Account.Columns().TenantId, tenantID).
		WhereIn(dao.Account.Columns().Number, numbers).
		Scan(&accounts)
	if err != nil {
		return nil, err
	}
	for _, account := range accounts {
		if account != nil {
			result[account.Number] = account
		}
	}
	return result, nil
}

func legacyAccountDetailsByIDs(ctx context.Context, tenantID int, accountIDs []int64) (map[int64]*entity.AccountDetail, error) {
	result := make(map[int64]*entity.AccountDetail, len(accountIDs))
	if len(accountIDs) == 0 {
		return result, nil
	}
	var details []*entity.AccountDetail
	err := dao.AccountDetail.Ctx(ctx).
		Where(dao.AccountDetail.Columns().TenantId, tenantID).
		WhereIn(dao.AccountDetail.Columns().AccountId, accountIDs).
		Scan(&details)
	if err != nil {
		return nil, err
	}
	for _, detail := range details {
		if detail != nil {
			result[detail.AccountId] = detail
		}
	}
	return result, nil
}

func legacyPhoneOwnersByPhone(ctx context.Context, tenantID int, phones []string) (map[string]map[int64]struct{}, error) {
	result := make(map[string]map[int64]struct{}, len(phones))
	if len(phones) == 0 {
		return result, nil
	}
	var accounts []*entity.Account
	err := dao.Account.Ctx(ctx).
		Unscoped().
		Fields(dao.Account.Columns().Id, dao.Account.Columns().Phone).
		Where(dao.Account.Columns().TenantId, tenantID).
		WhereIn(dao.Account.Columns().Phone, phones).
		Scan(&accounts)
	if err != nil {
		return nil, err
	}
	for _, account := range accounts {
		if account == nil || account.Phone == "" {
			continue
		}
		if result[account.Phone] == nil {
			result[account.Phone] = map[int64]struct{}{}
		}
		result[account.Phone][account.Id] = struct{}{}
	}
	return result, nil
}

func phoneUsedByOther(owners map[string]map[int64]struct{}, phone string, accountID int64) bool {
	ids := owners[phone]
	if len(ids) == 0 {
		return false
	}
	for id := range ids {
		if id != accountID {
			return true
		}
	}
	return false
}

func legacyUnitsByCode(ctx context.Context, tenantID int) (map[string]*entity.Unit, error) {
	var units []*entity.Unit
	err := dao.Unit.Ctx(ctx).
		Fields(dao.Unit.Columns().Id, dao.Unit.Columns().Code).
		Where(dao.Unit.Columns().TenantId, tenantID).
		Scan(&units)
	if err != nil {
		return nil, err
	}
	result := make(map[string]*entity.Unit, len(units))
	for _, unit := range units {
		if unit != nil && unit.Code != "" {
			result[unit.Code] = unit
		}
	}
	return result, nil
}

func unitID(units map[string]*entity.Unit, code string) int64 {
	if unit, ok := units[strings.TrimSpace(code)]; ok && unit != nil {
		return unit.Id
	}
	return 0
}

func detailDO(tenantID int, accountID int64, input accountDetailSyncInput, create bool) do.AccountDetail {
	data := do.AccountDetail{
		AccountId: accountID,
		TenantId:  tenantID,
		UpdatedBy: int64(0),
	}
	if create {
		data.CreatedBy = int64(0)
	}
	if input.source == "" {
		input.source = legacySourceSync
	}
	data.Source = input.source
	data.Idcard = input.idcard
	data.Birthday = input.birthday
	data.Email = input.email
	data.Gender = input.gender
	data.Qq = input.qq
	data.Wechat = input.wechat
	data.Avatar = input.avatar
	data.Grade = input.grade
	data.College = input.college
	data.CollegeCode = input.collegeCode
	data.Campus = input.campus
	data.SchoolSystem = input.schoolSystem
	data.GraduatedAt = input.graduatedAt
	data.Major = input.major
	data.ClassName = input.className
	data.Face = input.face
	return data
}

func changedDetailDO(detail *entity.AccountDetail, input accountDetailSyncInput) (do.AccountDetail, bool) {
	data := do.AccountDetail{UpdatedBy: int64(0)}
	changed := false
	setString := func(next string, current string, target *any) {
		if next != "" && next != current {
			*target = next
			changed = true
		}
	}
	setString(input.idcard, detail.Idcard, &data.Idcard)
	setString(input.birthday, detail.Birthday, &data.Birthday)
	setString(input.email, detail.Email, &data.Email)
	setString(input.qq, detail.Qq, &data.Qq)
	setString(input.wechat, detail.Wechat, &data.Wechat)
	setString(input.avatar, detail.Avatar, &data.Avatar)
	setString(input.grade, detail.Grade, &data.Grade)
	setString(input.college, detail.College, &data.College)
	setString(input.collegeCode, detail.CollegeCode, &data.CollegeCode)
	setString(input.campus, detail.Campus, &data.Campus)
	setString(input.schoolSystem, detail.SchoolSystem, &data.SchoolSystem)
	setString(input.graduatedAt, detail.GraduatedAt, &data.GraduatedAt)
	setString(input.major, detail.Major, &data.Major)
	setString(input.className, detail.ClassName, &data.ClassName)
	setString(input.face, detail.Face, &data.Face)
	if input.gender > 0 && input.gender != detail.Gender {
		data.Gender = input.gender
		changed = true
	}
	return data, changed
}

func mergeAccountAudit(account *entity.Account, data do.Account) map[string]any {
	result := accountAuditRecord(account)
	if result == nil {
		result = map[string]any{}
	}
	mergeAuditValue(result, "name", data.Name)
	mergeAuditValue(result, "phone", data.Phone)
	mergeAuditValue(result, "containerId", data.ContainerId)
	mergeAuditValue(result, "unitId", data.UnitId)
	mergeAuditValue(result, "status", data.Status)
	mergeAuditValue(result, "updatedBy", data.UpdatedBy)
	return result
}

func mergeDetailAudit(detail *entity.AccountDetail, data do.AccountDetail) map[string]any {
	result := detailAuditRecord(detail)
	if result == nil {
		result = map[string]any{}
	}
	mergeAuditValue(result, "birthday", data.Birthday)
	mergeAuditValue(result, "email", data.Email)
	mergeAuditValue(result, "gender", data.Gender)
	mergeAuditValue(result, "qq", data.Qq)
	mergeAuditValue(result, "wechat", data.Wechat)
	mergeAuditValue(result, "idcard", data.Idcard)
	mergeAuditValue(result, "avatar", data.Avatar)
	mergeAuditValue(result, "source", data.Source)
	mergeAuditValue(result, "grade", data.Grade)
	mergeAuditValue(result, "college", data.College)
	mergeAuditValue(result, "collegeCode", data.CollegeCode)
	mergeAuditValue(result, "campus", data.Campus)
	mergeAuditValue(result, "schoolSystem", data.SchoolSystem)
	mergeAuditValue(result, "graduatedAt", data.GraduatedAt)
	mergeAuditValue(result, "major", data.Major)
	mergeAuditValue(result, "className", data.ClassName)
	mergeAuditValue(result, "face", data.Face)
	mergeAuditValue(result, "updatedBy", data.UpdatedBy)
	return result
}

func mergeAuditValue(record map[string]any, key string, value any) {
	if value != nil {
		record[key] = value
	}
}

func sanitizeAccountInputs(inputs []*accountSyncInput) {
	for _, input := range inputs {
		if input == nil {
			continue
		}
		input.number = normalizeText(input.number)
		input.name = normalizeText(input.name)
		input.phone = normalizePhone(input.phone)
		input.statusRaw = normalizeText(input.statusRaw)
		input.unitCode = normalizeText(input.unitCode)
		input.containerName = normalizeText(input.containerName)
		input.groupName = normalizeText(input.groupName)
		input.detail.idcard = normalizeText(input.detail.idcard)
		input.detail.birthday = parseDateString(input.detail.birthday)
		input.detail.email = normalizeText(input.detail.email)
		input.detail.qq = normalizeText(input.detail.qq)
		input.detail.wechat = normalizeText(input.detail.wechat)
		input.detail.avatar = strings.TrimSpace(input.detail.avatar)
		input.detail.source = normalizeText(input.detail.source)
		input.detail.grade = normalizeText(input.detail.grade)
		input.detail.college = normalizeText(input.detail.college)
		input.detail.collegeCode = normalizeText(input.detail.collegeCode)
		input.detail.campus = normalizeText(input.detail.campus)
		input.detail.schoolSystem = normalizeText(input.detail.schoolSystem)
		input.detail.graduatedAt = normalizeText(input.detail.graduatedAt)
		input.detail.major = normalizeText(input.detail.major)
		input.detail.className = normalizeText(input.detail.className)
		input.detail.face = strings.TrimSpace(input.detail.face)
		if input.phone == "" {
			input.phone = input.number
		}
	}
}

func prepareAccountSyncInputs(inputs []*accountSyncInput) ([]*accountSyncInput, jobRunStats) {
	if len(inputs) == 0 {
		return nil, jobRunStats{}
	}
	sanitizeAccountInputs(inputs)
	stats := jobRunStats{}
	result := make([]*accountSyncInput, 0, len(inputs))
	for _, input := range inputs {
		if input == nil {
			continue
		}
		if input.number == "" {
			stats.errNum++
			continue
		}
		if input.requireNumericUnitCode && !legacyNumericUnitCode(input.unitCode) {
			stats.errNum++
			continue
		}
		result = append(result, input)
	}
	return result, stats
}

func legacyNumericUnitCode(value string) bool {
	_, err := strconv.Atoi(strings.TrimSpace(value))
	return err == nil
}

func unsupportedOracleDriverError(err error) error {
	if err == nil {
		return nil
	}
	return gerror.Wrap(err, "uidentity oracle source is unavailable")
}

func oraclePageModel(db gdb.DB, table string, page int, pageSize int) *gdb.Model {
	return db.Model(table).Safe().Offset(page * pageSize).Limit(pageSize)
}

func sqlLogFields(stats jobRunStats) g.Map {
	return g.Map{
		"createNum":              stats.createNum,
		"updateNum":              stats.updateNum,
		"updateAccountCount":     stats.updateAccountCount,
		"updateAccountDetailNum": stats.updateAccountDetailCount,
		"deleteNum":              stats.deleteNum,
		"errNum":                 stats.errNum,
	}
}
