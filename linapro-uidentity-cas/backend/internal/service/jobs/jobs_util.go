// This file contains shared helpers for UIdentity scheduled-job handlers,
// including tenant resolution, external configuration, normalization, and audit
// writes for task-driven account mutations.

package jobs

import (
	"context"
	"crypto/rand"
	"crypto/sha1"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"regexp"
	"slices"
	"strings"
	"time"

	"github.com/gogf/gf/v2/database/gdb"

	"lina-core/pkg/bizerr"
	plugincontract "lina-core/pkg/plugin/capability/contract"
	"lina-plugin-linapro-uidentity-cas/backend/internal/dao"
	"lina-plugin-linapro-uidentity-cas/backend/internal/model/do"
	"lina-plugin-linapro-uidentity-cas/backend/internal/model/entity"
)

const (
	legacySourceSync = "sync"

	legacyContainerStudent   = "bzks"
	legacyContainerStudentWJ = "wjs"
	legacyContainerStudentYJ = "yjs"
	legacyContainerStaff     = "jzg"
	legacyContainerAlumni    = "xy"

	legacyAuditTableAccount = "account"
	legacyAuditTableDetail  = "account_details"
	legacyAuditCreate       = "create"
	legacyAuditUpdate       = "update"

	defaultPageSize            = 1000
	defaultLDAPSearchPageSize  = 100
	defaultLDAPPassword        = "Isicau@2024"
	defaultLDAPPeopleOU        = "People"
	defaultLDAPStaffOU         = "jzg"
	defaultLDAPActiveStatus    = "active"
	defaultLegacyExpireAt      = "2099-12-31"
	configKeyOracleDSN         = "legacy.jobs.oracle.dsn"
	configKeyNewUserTime       = "legacy.jobs.newUserTime"
	configKeyLDAPAddr          = "legacy.jobs.ldap.addr"
	configKeyLDAPBindDN        = "legacy.jobs.ldap.bindDN"
	configKeyLDAPBindPassword  = "legacy.jobs.ldap.bindPassword"
	configKeyLDAPBaseDN        = "legacy.jobs.ldap.baseDN"
	configKeyLDAPObjectClass   = "legacy.jobs.ldap.objectClass"
	configKeyLDAPDefaultPass   = "legacy.jobs.ldap.defaultPassword"
	configKeyLDAPSyncPageSize  = "legacy.jobs.ldap.syncPageSize"
	configKeyLDAPSkipTLSVerify = "legacy.jobs.ldap.skipTLSVerify"
)

var phonePattern = regexp.MustCompile(`1[3-9][0-9]{9}`)

func (s *serviceImpl) tenantID(ctx context.Context) int {
	if s != nil && s.bizCtxSvc != nil {
		current := s.bizCtxSvc.Current(ctx)
		return current.TenantID
	}
	return plugincontract.CurrentFromContext(ctx).TenantID
}

func (s *serviceImpl) requireConfigString(ctx context.Context, key string) (string, error) {
	if s == nil || s.configSvc == nil {
		return "", pluginJobUnsupported()
	}
	value, err := s.configSvc.String(ctx, key, "")
	if err != nil {
		return "", err
	}
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return "", pluginJobUnsupported()
	}
	return trimmed, nil
}

func (s *serviceImpl) configString(ctx context.Context, key string, defaultValue string) (string, error) {
	if s == nil || s.configSvc == nil {
		return defaultValue, nil
	}
	value, err := s.configSvc.String(ctx, key, defaultValue)
	if err != nil {
		return "", err
	}
	if strings.TrimSpace(value) == "" {
		return defaultValue, nil
	}
	return strings.TrimSpace(value), nil
}

func (s *serviceImpl) configInt(ctx context.Context, key string, defaultValue int) (int, error) {
	if s == nil || s.configSvc == nil {
		return defaultValue, nil
	}
	value, err := s.configSvc.Int(ctx, key, defaultValue)
	if err != nil {
		return 0, err
	}
	if value <= 0 {
		return defaultValue, nil
	}
	return value, nil
}

func (s *serviceImpl) newUserCutoff(ctx context.Context) (time.Time, error) {
	value, err := s.configString(ctx, configKeyNewUserTime, "")
	if err != nil || strings.TrimSpace(value) == "" {
		return time.Time{}, err
	}
	for _, layout := range []string{time.DateOnly, time.DateTime, time.RFC3339} {
		parsed, parseErr := time.ParseInLocation(layout, value, time.Local)
		if parseErr == nil {
			return parsed, nil
		}
	}
	return time.Time{}, nil
}

func pluginJobUnsupported() error {
	return bizerr.NewCode(CodeJobExecutorUnsupported)
}

func normalizeText(value string) string {
	return strings.Join(strings.Fields(strings.TrimSpace(value)), "")
}

func normalizePhone(value string) string {
	normalized := normalizeText(value)
	if match := phonePattern.FindString(normalized); match != "" {
		return match
	}
	return normalized
}

func genderValue(value string) int {
	switch strings.TrimSpace(value) {
	case "男", "1":
		return 1
	case "女", "2":
		return 2
	default:
		return 0
	}
}

func statusValue(raw string, create bool, newUserCutoff time.Time) int {
	switch strings.TrimSpace(raw) {
	case "待报到":
		return 0
	case "在读", "参军", "休学", "交换期溝", "保留学籍", "on", "2", "在籍", "1":
		if create && time.Now().After(newUserCutoff) {
			return 0
		}
		return 1
	case "保留入学资格", "结业", "毕业", "转学(转出)", "退学", "未在籍", "quit", "die":
		return 2
	default:
		return 0
	}
}

func parseDateString(value string) string {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return ""
	}
	for _, layout := range []string{time.DateOnly, time.DateTime, "20060102"} {
		parsed, err := time.ParseInLocation(layout, trimmed, time.Local)
		if err == nil {
			return parsed.Format(time.DateOnly)
		}
	}
	return ""
}

func birthdayFromIDCard(idcard string) string {
	trimmed := strings.TrimSpace(idcard)
	if len(trimmed) == 18 {
		return parseDateString(trimmed[6:14])
	}
	return parseDateString("19900101")
}

func defaultExpireAt() *time.Time {
	expireAt, err := time.ParseInLocation(time.DateOnly, defaultLegacyExpireAt, time.Local)
	if err != nil {
		return nil
	}
	expireAt = expireAt.Add(23*time.Hour + 59*time.Minute + 59*time.Second)
	return &expireAt
}

func sleepWithContext(ctx context.Context, duration time.Duration) bool {
	timer := time.NewTimer(duration)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return false
	case <-timer.C:
		return true
	}
}

func idsFromRecords(accounts []*entity.Account) []int64 {
	ids := make([]int64, 0, len(accounts))
	for _, account := range accounts {
		if account != nil && account.Id > 0 {
			ids = append(ids, account.Id)
		}
	}
	return ids
}

func stringsFromSet(values map[string]struct{}) []string {
	result := make([]string, 0, len(values))
	for value := range values {
		if strings.TrimSpace(value) != "" {
			result = append(result, value)
		}
	}
	slices.Sort(result)
	return result
}

func int64sFromSet(values map[int64]struct{}) []int64 {
	result := make([]int64, 0, len(values))
	for value := range values {
		if value > 0 {
			result = append(result, value)
		}
	}
	slices.Sort(result)
	return result
}

func (s *serviceImpl) insertAccountAudit(ctx context.Context, tenantID int, accountID int64, tableName string, action string, oldData any, newData any) error {
	_, err := dao.AccountChangeLog.Ctx(ctx).Data(do.AccountChangeLog{
		TenantId:  tenantID,
		AccountId: accountID,
		TableName: tableName,
		Action:    action,
		DataOld:   auditJSON(oldData),
		DataNew:   auditJSON(newData),
	}).Insert()
	return err
}

func auditJSON(value any) string {
	if value == nil {
		return ""
	}
	data, err := json.Marshal(value)
	if err != nil {
		return ""
	}
	return string(data)
}

func accountAuditRecord(account *entity.Account) map[string]any {
	if account == nil {
		return nil
	}
	return map[string]any{
		"id":                account.Id,
		"tenantId":          account.TenantId,
		"number":            account.Number,
		"name":              account.Name,
		"phone":             account.Phone,
		"effectAt":          account.EffectAt,
		"expireAt":          account.ExpireAt,
		"passwordUpdatedAt": account.PasswordUpdatedAt,
		"passLevel":         account.PassLevel,
		"containerId":       account.ContainerId,
		"unitId":            account.UnitId,
		"status":            account.Status,
		"createdBy":         account.CreatedBy,
		"updatedBy":         account.UpdatedBy,
		"createdAt":         account.CreatedAt,
		"updatedAt":         account.UpdatedAt,
		"deletedAt":         account.DeletedAt,
	}
}

func detailAuditRecord(detail *entity.AccountDetail) map[string]any {
	if detail == nil {
		return nil
	}
	return map[string]any{
		"accountId":    detail.AccountId,
		"tenantId":     detail.TenantId,
		"birthday":     detail.Birthday,
		"email":        detail.Email,
		"gender":       detail.Gender,
		"qq":           detail.Qq,
		"wechat":       detail.Wechat,
		"idcard":       detail.Idcard,
		"avatar":       detail.Avatar,
		"source":       detail.Source,
		"grade":        detail.Grade,
		"college":      detail.College,
		"collegeCode":  detail.CollegeCode,
		"campus":       detail.Campus,
		"schoolSystem": detail.SchoolSystem,
		"graduatedAt":  detail.GraduatedAt,
		"major":        detail.Major,
		"className":    detail.ClassName,
		"face":         detail.Face,
		"createdBy":    detail.CreatedBy,
		"updatedBy":    detail.UpdatedBy,
		"createdAt":    detail.CreatedAt,
		"updatedAt":    detail.UpdatedAt,
	}
}

func hashAccountPassword(password string) string {
	sum := sha256.Sum256([]byte(password))
	return hex.EncodeToString(sum[:])
}

func generateLDAPSSHA(password string) (string, error) {
	salt := make([]byte, 8)
	if _, err := rand.Read(salt); err != nil {
		return "", err
	}
	hash := sha1.New()
	if _, err := hash.Write([]byte(password)); err != nil {
		return "", err
	}
	if _, err := hash.Write(salt); err != nil {
		return "", err
	}
	ssha := append(hash.Sum(nil), salt...)
	return "{SSHA}" + base64.StdEncoding.EncodeToString(ssha), nil
}

func scanAllPages[T any](ctx context.Context, model *gdb.Model, pageSize int, handle func([]*T) error) error {
	if pageSize <= 0 {
		pageSize = defaultPageSize
	}
	for page := 0; ; page++ {
		var rows []*T
		if err := model.Offset(page * pageSize).Limit(pageSize).Scan(&rows); err != nil {
			return err
		}
		if len(rows) == 0 {
			return nil
		}
		if err := handle(rows); err != nil {
			return err
		}
		if len(rows) < pageSize {
			return nil
		}
	}
}
