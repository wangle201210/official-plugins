// This file implements resource registry, paged reads, generic CRUD dispatch,
// API field projection, and tenant-scoped delete validation for plugin tables.

package uidentity

import (
	"context"
	"strings"
	"time"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/util/gconv"

	"lina-core/pkg/apitime"
	"lina-core/pkg/bizerr"
	"lina-plugin-linapro-uidentity-cas/backend/internal/dao"
)

const (
	maxDeleteIDs = 100
)

type resourceDefinition struct {
	name          string
	table         string
	idColumn      string
	defaultOrder  string
	keywordFields []string
	apiToColumn   map[string]string
	timeFields    map[string]struct{}
	dateFields    map[string]struct{}
	model         func(context.Context) *gdb.Model
	data          func(context.Context, map[string]any, bool) (any, error)
}

func (s *serviceImpl) resourceDefinition(resource string) (*resourceDefinition, error) {
	definitions := s.resourceDefinitions()
	if def, ok := definitions[strings.TrimSpace(resource)]; ok {
		return def, nil
	}
	return nil, bizerr.NewCode(CodeResourceNotSupported)
}

func (s *serviceImpl) resourceDefinitions() map[string]*resourceDefinition {
	return map[string]*resourceDefinition{
		"accounts":               s.accountResource(),
		"account-details":        s.accountDetailResource(),
		"groups":                 s.groupResource(),
		"units":                  s.unitResource(),
		"containers":             s.containerResource(),
		"applications":           s.applicationResource(),
		"account-groups":         s.accountGroupResource(),
		"account-units":          s.accountUnitResource(),
		"account-app-roles":      s.accountAppRoleResource(),
		"account-app-blacklists": s.accountAppBlacklistResource(),
		"group-app-blacklists":   s.groupAppBlacklistResource(),
		"pass-rules":             s.passRuleResource(),
		"sms-records":            s.smsResource(),
		"cas-login-logs":         s.casLoginLogResource(),
		"oauth-logs":             s.oauthLogResource(),
		"oauth-tokens":           s.oauthTokenResource(),
		"account-change-logs":    s.accountChangeLogResource(),
		"account-app-blacklist":  s.accountAppBlacklistResource(),
		"group-app-blacklist":    s.groupAppBlacklistResource(),
		"account-app-role":       s.accountAppRoleResource(),
		"account-change-log":     s.accountChangeLogResource(),
		"account-details-legacy": s.accountDetailResource(),
		"cas-login-logs-legacy":  s.casLoginLogResource(),
		"oauth-log":              s.oauthLogResource(),
		"oauth-token":            s.oauthTokenResource(),
		"pass-ruler":             s.passRuleResource(),
		"sms":                    s.smsResource(),
		"account-unit":           s.accountUnitResource(),
	}
}

// ListResource returns one paged tenant-scoped resource list.
func (s *serviceImpl) ListResource(ctx context.Context, in ResourceListInput) (*ResourceListOutput, error) {
	def, err := s.resourceDefinition(in.Resource)
	if err != nil {
		return nil, err
	}
	model := s.applyResourceFilters(ctx, def, in)
	total, err := model.Count()
	if err != nil {
		return nil, err
	}
	orderColumn := def.defaultOrder
	if column := def.apiToColumn[strings.TrimSpace(in.OrderBy)]; column != "" {
		orderColumn = column
	}
	if strings.EqualFold(in.Order, "asc") {
		model = model.OrderAsc(orderColumn)
	} else {
		model = model.OrderDesc(orderColumn)
	}
	result, err := model.
		Fields(projectionFields(def)...).
		Page(in.PageNum, in.PageSize).
		All()
	if err != nil {
		return nil, err
	}
	records := projectResult(result, def)
	if def.name == "accounts" && len(records) > 0 {
		if err := s.decorateAccountRecords(ctx, records); err != nil {
			return nil, err
		}
	}
	return &ResourceListOutput{List: records, Total: total}, nil
}

// GetResource returns one tenant-scoped resource detail.
func (s *serviceImpl) GetResource(ctx context.Context, resource string, id int64) (Record, error) {
	def, err := s.resourceDefinition(resource)
	if err != nil {
		return nil, err
	}
	result, err := s.tenantFilter.Apply(ctx, def.model(ctx), "").
		Fields(projectionFields(def)...).
		Where(def.idColumn, id).
		One()
	if err != nil {
		return nil, err
	}
	if result.IsEmpty() {
		return nil, bizerr.NewCode(CodeResourceNotFound)
	}
	record := projectRecord(result, def)
	if def.name == "accounts" {
		if err := s.decorateAccountRecords(ctx, []Record{record}); err != nil {
			return nil, err
		}
	}
	return record, nil
}

// CreateResource creates one tenant-scoped resource row.
func (s *serviceImpl) CreateResource(ctx context.Context, resource string, body map[string]any) (int64, error) {
	def, err := s.resourceDefinition(resource)
	if err != nil {
		return 0, err
	}
	data, err := def.data(ctx, body, true)
	if err != nil {
		return 0, err
	}
	if def.name == "account-details" {
		_, err := def.model(ctx).Data(data).Insert()
		if err != nil {
			return 0, err
		}
		return int64Field(body, "accountId"), nil
	}
	id, err := def.model(ctx).Data(data).InsertAndGetId()
	if err != nil {
		return 0, err
	}
	if def.name == "accounts" {
		if err := s.ensureAccountDetail(ctx, id); err != nil {
			return 0, err
		}
	}
	return id, nil
}

// UpdateResource updates one tenant-scoped resource row.
func (s *serviceImpl) UpdateResource(ctx context.Context, resource string, id int64, body map[string]any) error {
	def, err := s.resourceDefinition(resource)
	if err != nil {
		return err
	}
	if _, err = s.GetResource(ctx, resource, id); err != nil {
		return err
	}
	data, err := def.data(ctx, body, false)
	if err != nil {
		return err
	}
	_, err = s.tenantFilter.Apply(ctx, def.model(ctx), "").
		Where(def.idColumn, id).
		OmitNilData().
		Data(data).
		Update()
	return err
}

// DeleteResource deletes one or more tenant-scoped resource rows.
func (s *serviceImpl) DeleteResource(ctx context.Context, resource string, ids string) error {
	def, err := s.resourceDefinition(resource)
	if err != nil {
		return err
	}
	idList := parseIDList(ids)
	if len(idList) == 0 {
		return bizerr.NewCode(CodeDeleteIDsRequired)
	}
	if len(idList) > maxDeleteIDs {
		return bizerr.NewCode(CodeDeleteIDsTooMany, bizerr.P("limit", maxDeleteIDs))
	}
	count, err := s.tenantFilter.Apply(ctx, def.model(ctx), "").
		WhereIn(def.idColumn, idList).
		Count()
	if err != nil {
		return err
	}
	if count != len(idList) {
		return bizerr.NewCode(CodeResourceNotFound)
	}
	_, err = s.tenantFilter.Apply(ctx, def.model(ctx), "").
		WhereIn(def.idColumn, idList).
		Delete()
	return err
}

func (s *serviceImpl) applyResourceFilters(ctx context.Context, def *resourceDefinition, in ResourceListInput) *gdb.Model {
	model := s.tenantFilter.Apply(ctx, def.model(ctx), "")
	keyword := strings.TrimSpace(in.Keyword)
	if keyword != "" && len(def.keywordFields) > 0 {
		likeConditions := make([]string, 0, len(def.keywordFields))
		likeValues := make([]any, 0, len(def.keywordFields))
		for _, field := range def.keywordFields {
			likeConditions = append(likeConditions, field+" LIKE ?")
			likeValues = append(likeValues, "%"+keyword+"%")
		}
		model = model.Where("("+strings.Join(likeConditions, " OR ")+")", likeValues...)
	}
	if in.AccountId > 0 && def.apiToColumn["accountId"] != "" {
		model = model.Where(def.apiToColumn["accountId"], in.AccountId)
	}
	if in.AppId > 0 && def.apiToColumn["appId"] != "" {
		model = model.Where(def.apiToColumn["appId"], in.AppId)
	}
	if in.GroupId > 0 && def.apiToColumn["groupId"] != "" {
		model = model.Where(def.apiToColumn["groupId"], in.GroupId)
	}
	if in.ContainerId > 0 && def.apiToColumn["containerId"] != "" {
		model = model.Where(def.apiToColumn["containerId"], in.ContainerId)
	}
	if in.UnitId > 0 && def.apiToColumn["unitId"] != "" {
		model = model.Where(def.apiToColumn["unitId"], in.UnitId)
	}
	if in.Status != nil && def.apiToColumn["status"] != "" {
		model = model.Where(def.apiToColumn["status"], *in.Status)
	}
	if len(in.PassLevels) > 0 && def.apiToColumn["passLevel"] != "" {
		model = model.WhereIn(def.apiToColumn["passLevel"], in.PassLevels)
	}
	if len(in.GroupIds) > 0 && def.name == "accounts" {
		groupColumns := dao.AccountGroup.Columns()
		subQuery := s.tenantFilter.Apply(ctx, dao.AccountGroup.Ctx(ctx), "").
			Fields(groupColumns.AccountId).
			WhereIn(groupColumns.GroupId, in.GroupIds)
		model = model.Where(def.idColumn+" IN (?)", subQuery)
	}
	return model
}

func projectResult(result gdb.Result, def *resourceDefinition) []Record {
	list := make([]Record, 0, len(result))
	for _, row := range result {
		list = append(list, projectRecord(row, def))
	}
	return list
}

func projectionColumns(def *resourceDefinition) []string {
	seen := make(map[string]struct{}, len(def.apiToColumn))
	columns := make([]string, 0, len(def.apiToColumn))
	for _, column := range def.apiToColumn {
		if column == "" {
			continue
		}
		if _, ok := seen[column]; ok {
			continue
		}
		columns = append(columns, column)
		seen[column] = struct{}{}
	}
	return columns
}

func projectionFields(def *resourceDefinition) []any {
	columns := projectionColumns(def)
	fields := make([]any, 0, len(columns))
	for _, column := range columns {
		fields = append(fields, column)
	}
	return fields
}

func projectRecord(row gdb.Record, def *resourceDefinition) Record {
	record := make(Record, len(def.apiToColumn))
	for apiName, columnName := range def.apiToColumn {
		value := row[columnName]
		if _, ok := def.timeFields[apiName]; ok {
			record[apiName] = apitime.MilliFromTime(value.Time())
			continue
		}
		if _, ok := def.dateFields[apiName]; ok {
			record[apiName] = value.String()
			continue
		}
		record[apiName] = value.Interface()
	}
	return record
}

func parseIDList(ids string) []int64 {
	parts := strings.Split(ids, ",")
	result := make([]int64, 0, len(parts))
	for _, part := range parts {
		id := gconv.Int64(strings.TrimSpace(part))
		if id > 0 {
			result = append(result, id)
		}
	}
	return result
}

func (s *serviceImpl) tenantID(ctx context.Context) int {
	return s.tenantFilter.Context(ctx).TenantID
}

func (s *serviceImpl) actorID(ctx context.Context) int64 {
	tenantCtx := s.tenantFilter.Context(ctx)
	if tenantCtx.ActingUserID > 0 {
		return int64(tenantCtx.ActingUserID)
	}
	return int64(tenantCtx.UserID)
}

func (s *serviceImpl) baseOwnedDO(ctx context.Context, create bool) (tenantID int, actorID int64) {
	return s.tenantID(ctx), s.actorID(ctx)
}

func hasField(body map[string]any, field string) bool {
	_, ok := body[field]
	return ok
}

func stringField(body map[string]any, field string) string {
	return strings.TrimSpace(gconv.String(body[field]))
}

func intField(body map[string]any, field string) int {
	return gconv.Int(body[field])
}

func int64Field(body map[string]any, field string) int64 {
	return gconv.Int64(body[field])
}

func timeField(body map[string]any, field string) *time.Time {
	if !hasField(body, field) {
		return nil
	}
	value := gconv.Time(body[field])
	if value.IsZero() {
		return nil
	}
	return &value
}

func mergeBody(reqBody map[string]any) map[string]any {
	if reqBody == nil {
		return map[string]any{}
	}
	return reqBody
}

func commonTimeFields() map[string]struct{} {
	return map[string]struct{}{
		"effectAt":          {},
		"expireAt":          {},
		"passwordUpdatedAt": {},
		"createdAt":         {},
		"updatedAt":         {},
		"deletedAt":         {},
		"loginTime":         {},
		"expiredAt":         {},
	}
}

func (s *serviceImpl) accountResource() *resourceDefinition {
	cols := dao.Account.Columns()
	return &resourceDefinition{
		name:          "accounts",
		table:         dao.Account.Table(),
		idColumn:      cols.Id,
		defaultOrder:  cols.Id,
		keywordFields: []string{cols.Number, cols.Name, cols.Phone},
		apiToColumn: map[string]string{
			"id": cols.Id, "tenantId": cols.TenantId, "number": cols.Number, "name": cols.Name, "phone": cols.Phone,
			"effectAt": cols.EffectAt, "expireAt": cols.ExpireAt, "passwordUpdatedAt": cols.PasswordUpdatedAt,
			"passLevel": cols.PassLevel, "containerId": cols.ContainerId, "unitId": cols.UnitId, "status": cols.Status,
			"createdBy": cols.CreatedBy, "updatedBy": cols.UpdatedBy, "createdAt": cols.CreatedAt, "updatedAt": cols.UpdatedAt, "deletedAt": cols.DeletedAt,
		},
		timeFields: commonTimeFields(),
		model:      func(ctx context.Context) *gdb.Model { return dao.Account.Ctx(ctx) },
		data:       s.accountData,
	}
}

func (s *serviceImpl) accountDetailResource() *resourceDefinition {
	cols := dao.AccountDetail.Columns()
	return &resourceDefinition{
		name:          "account-details",
		table:         dao.AccountDetail.Table(),
		idColumn:      cols.AccountId,
		defaultOrder:  cols.AccountId,
		keywordFields: []string{cols.Email, cols.Wechat, cols.Idcard},
		apiToColumn: map[string]string{
			"accountId": cols.AccountId, "tenantId": cols.TenantId, "birthday": cols.Birthday, "email": cols.Email, "gender": cols.Gender,
			"qq": cols.Qq, "wechat": cols.Wechat, "idcard": cols.Idcard, "avatar": cols.Avatar, "source": cols.Source,
			"grade": cols.Grade, "college": cols.College, "collegeCode": cols.CollegeCode, "campus": cols.Campus,
			"schoolSystem": cols.SchoolSystem, "graduatedAt": cols.GraduatedAt, "major": cols.Major, "className": cols.ClassName,
			"face": cols.Face, "createdBy": cols.CreatedBy, "updatedBy": cols.UpdatedBy, "createdAt": cols.CreatedAt, "updatedAt": cols.UpdatedAt,
		},
		timeFields: commonTimeFields(),
		dateFields: map[string]struct{}{"birthday": {}},
		model:      func(ctx context.Context) *gdb.Model { return dao.AccountDetail.Ctx(ctx) },
		data:       s.accountDetailData,
	}
}

func (s *serviceImpl) groupResource() *resourceDefinition {
	cols := dao.Group.Columns()
	return simpleNamedResource("groups", dao.Group.Table(), cols.Id, cols.Name, cols.Alias, cols.CreatedBy, cols.UpdatedBy, cols.CreatedAt, cols.UpdatedAt, cols.DeletedAt, func(ctx context.Context) *gdb.Model { return dao.Group.Ctx(ctx) }, s.groupData)
}

func (s *serviceImpl) unitResource() *resourceDefinition {
	cols := dao.Unit.Columns()
	def := simpleNamedResource("units", dao.Unit.Table(), cols.Id, cols.Name, cols.Alias, cols.CreatedBy, cols.UpdatedBy, cols.CreatedAt, cols.UpdatedAt, cols.DeletedAt, func(ctx context.Context) *gdb.Model { return dao.Unit.Ctx(ctx) }, s.unitData)
	def.apiToColumn["code"] = cols.Code
	def.keywordFields = append(def.keywordFields, cols.Code)
	return def
}

func (s *serviceImpl) containerResource() *resourceDefinition {
	cols := dao.Container.Columns()
	def := simpleNamedResource("containers", dao.Container.Table(), cols.Id, cols.Name, cols.Alias, cols.CreatedBy, cols.UpdatedBy, cols.CreatedAt, cols.UpdatedAt, cols.DeletedAt, func(ctx context.Context) *gdb.Model { return dao.Container.Ctx(ctx) }, s.containerData)
	def.apiToColumn["accountCount"] = cols.AccountCount
	def.apiToColumn["adminCount"] = cols.AdminCount
	return def
}

func simpleNamedResource(name, table, id, resourceName, alias, createdBy, updatedBy, createdAt, updatedAt, deletedAt string, model func(context.Context) *gdb.Model, data func(context.Context, map[string]any, bool) (any, error)) *resourceDefinition {
	return &resourceDefinition{
		name:          name,
		table:         table,
		idColumn:      id,
		defaultOrder:  id,
		keywordFields: []string{resourceName, alias},
		apiToColumn: map[string]string{
			"id": id, "name": resourceName, "alias": alias, "createdBy": createdBy, "updatedBy": updatedBy,
			"createdAt": createdAt, "updatedAt": updatedAt, "deletedAt": deletedAt,
		},
		timeFields: commonTimeFields(),
		model:      model,
		data:       data,
	}
}

// The remaining resource definitions are placed in uidentity_resource_defs.go
// to keep this registry file readable.
