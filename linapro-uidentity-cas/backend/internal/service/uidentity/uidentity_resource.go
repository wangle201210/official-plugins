// This file implements resource registry, paged reads, generic CRUD dispatch,
// API field projection, and delete validation for plugin tables.

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
	likeFields    map[string]struct{}
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
		"account-active-logs":    s.accountActiveLogResource(),
		"account-app-blacklist":  s.accountAppBlacklistResource(),
		"group-app-blacklist":    s.groupAppBlacklistResource(),
		"account-app-role":       s.accountAppRoleResource(),
		"account-change-log":     s.accountChangeLogResource(),
		"account-active-log":     s.accountActiveLogResource(),
		"account-details-legacy": s.accountDetailResource(),
		"cas-login-logs-legacy":  s.casLoginLogResource(),
		"oauth-log":              s.oauthLogResource(),
		"oauth-token":            s.oauthTokenResource(),
		"pass-ruler":             s.passRuleResource(),
		"sms":                    s.smsResource(),
		"account-unit":           s.accountUnitResource(),
	}
}

// ListResource returns one paged global resource list.
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
	} else if apiName := resourceFilterAPIName(in.OrderBy); apiName != "" {
		orderColumn = def.apiToColumn[apiName]
	}
	if orderColumn == "" {
		orderColumn = def.defaultOrder
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

// GetResource returns one global resource detail.
func (s *serviceImpl) GetResource(ctx context.Context, resource string, id int64) (Record, error) {
	def, err := s.resourceDefinition(resource)
	if err != nil {
		return nil, err
	}
	result, err := def.model(ctx).
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

// CreateResource creates one global resource row.
func (s *serviceImpl) CreateResource(ctx context.Context, resource string, body map[string]any) (int64, error) {
	body = normalizeLegacyResourceBody(resource, body)
	def, err := s.resourceDefinition(resource)
	if err != nil {
		return 0, err
	}
	data, err := def.data(ctx, body, true)
	if err != nil {
		return 0, err
	}
	if def.name == "account-details" {
		accountID := int64Field(body, "accountId")
		if err := s.createAccountDetailWithAudit(ctx, data, accountID); err != nil {
			return 0, err
		}
		return accountID, nil
	}
	if def.name == "accounts" {
		if groupIDs, ok := accountGroupIDsFromBody(body); ok {
			if err := s.validateAccountGroups(ctx, groupIDs); err != nil {
				return 0, err
			}
		}
		id, err := s.createAccountWithAudit(ctx, data)
		if err != nil {
			return 0, err
		}
		if err := s.ensureAccountDetail(ctx, id); err != nil {
			return 0, err
		}
		if err := s.syncAccountGroupsFromBody(ctx, id, body); err != nil {
			return 0, err
		}
		return id, nil
	}
	id, err := def.model(ctx).Data(data).InsertAndGetId()
	if err != nil {
		return 0, err
	}
	return id, nil
}

// UpdateResource updates one global resource row.
func (s *serviceImpl) UpdateResource(ctx context.Context, resource string, id int64, body map[string]any) error {
	body = normalizeLegacyResourceBody(resource, body)
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
	if def.name == "accounts" {
		if err := s.updateAccountWithAudit(ctx, id, data); err != nil {
			return err
		}
		return s.syncAccountGroupsFromBody(ctx, id, body)
	}
	if def.name == "account-details" {
		return s.updateAccountDetailWithAudit(ctx, id, data)
	}
	_, err = def.model(ctx).
		Where(def.idColumn, id).
		OmitNilData().
		Data(data).
		Update()
	return err
}

// DeleteResource deletes one or more global resource rows.
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
	count, err := def.model(ctx).
		WhereIn(def.idColumn, idList).
		Count()
	if err != nil {
		return err
	}
	if count != len(idList) {
		return bizerr.NewCode(CodeResourceNotFound)
	}
	return s.deleteResourceWithAccountAudit(ctx, def, idList)
}

func (s *serviceImpl) applyResourceFilters(ctx context.Context, def *resourceDefinition, in ResourceListInput) *gdb.Model {
	model := def.model(ctx)
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
	model = applyLegacyResourceFieldFilters(model, def, in)
	if len(in.PassLevels) > 0 && def.apiToColumn["passLevel"] != "" {
		model = model.WhereIn(def.apiToColumn["passLevel"], in.PassLevels)
	}
	if len(in.GroupIds) > 0 && def.name == "accounts" {
		groupColumns := dao.AccountGroup.Columns()
		subQuery := dao.AccountGroup.Ctx(ctx).
			Fields(groupColumns.AccountId).
			WhereIn(groupColumns.GroupsId, in.GroupIds)
		model = model.Where(def.idColumn+" IN (?)", subQuery)
	}
	return model
}

func applyLegacyResourceFieldFilters(model *gdb.Model, def *resourceDefinition, in ResourceListInput) *gdb.Model {
	if len(in.Filters) == 0 {
		return model
	}
	for fieldName, value := range in.Filters {
		apiName := resourceFilterAPIName(fieldName)
		column := def.apiToColumn[apiName]
		if column == "" || isResourceListControlField(apiName) || isResourceListFilterAlreadyApplied(apiName) {
			continue
		}
		if strings.TrimSpace(gconv.String(value)) == "" {
			continue
		}
		if _, ok := def.likeFields[apiName]; ok {
			model = model.Where(column+" LIKE ?", "%"+strings.TrimSpace(gconv.String(value))+"%")
			continue
		}
		model = model.Where(column, value)
	}
	return model
}

func resourceFilterAPIName(name string) string {
	trimmed := strings.TrimSpace(name)
	if trimmed == "" {
		return ""
	}
	return resourceLegacyFieldAliases()[trimmed]
}

func resourceLegacyFieldAliases() map[string]string {
	return map[string]string{
		"id": "id", "tenantId": "tenantId", "tenant_id": "tenantId",
		"accountId": "accountId", "account_id": "accountId",
		"appId": "appId", "app_id": "appId", "applicationId": "appId", "application_id": "appId",
		"groupId": "groupId", "group_id": "groupId",
		"containerId": "containerId", "container_id": "containerId",
		"unitId": "unitId", "unit_id": "unitId",
		"number": "number", "name": "name", "phone": "phone", "status": "status",
		"alias": "alias", "code": "code", "accountCount": "accountCount", "account_count": "accountCount",
		"adminCount": "adminCount", "admin_count": "adminCount",
		"birthday": "birthday", "email": "email", "gender": "gender", "qq": "qq", "wechat": "wechat",
		"idcard": "idcard", "idCard": "idcard", "avatar": "avatar", "source": "source",
		"grade": "grade", "nj": "grade", "college": "college", "xymc": "college",
		"collegeCode": "collegeCode", "college_code": "collegeCode", "xydm": "collegeCode",
		"campus": "campus", "xq": "campus", "schoolSystem": "schoolSystem", "school_system": "schoolSystem", "xz": "schoolSystem",
		"graduatedAt": "graduatedAt", "graduated_at": "graduatedAt", "yjbysj": "graduatedAt",
		"major": "major", "zymc": "major", "className": "className", "class_name": "className", "bjmc": "className",
		"face": "face", "giveAccountId": "giveAccountId", "give_account_id": "giveAccountId",
		"empoweredAccountId": "empoweredAccountId", "empowered_account_id": "empoweredAccountId",
		"effectAt": "effectAt", "effect_at": "effectAt", "expireAt": "expireAt", "expire_at": "expireAt",
		"clientId": "clientId", "client_id": "clientId", "secretKey": "secretKey", "secret_key": "secretKey",
		"accessModel": "accessModel", "access_model": "accessModel", "callbackUrl": "callbackUrl", "callback_url": "callbackUrl",
		"whitelist": "whitelist", "capital": "capital", "lower": "lower", "symbol": "symbol", "length": "length",
		"interval": "intervalDays", "intervalDays": "intervalDays", "interval_days": "intervalDays",
		"intervalStatus": "intervalStatus", "interval_status": "intervalStatus",
		"type": "type", "content": "content", "respMsg": "respMsg", "resp_msg": "respMsg",
		"choiceAccountId": "choiceAccountId", "choice_account_id": "choiceAccountId",
		"ipaddr": "ipaddr", "loginLocation": "loginLocation", "login_location": "loginLocation",
		"browser": "browser", "os": "os", "platform": "platform", "loginTime": "loginTime", "login_time": "loginTime",
		"remark": "remark", "msg": "msg", "loginType": "loginType", "login_type": "loginType",
		"userId": "userId", "user_id": "userId", "redirectUri": "redirectUri", "redirect_uri": "redirectUri", "scope": "scope",
		"expiredAt": "expiredAt", "expired_at": "expiredAt", "access": "access", "refresh": "refresh", "data": "data",
		"tableName": "tableName", "table_name": "tableName", "action": "action", "dataOld": "dataOld", "data_old": "dataOld",
		"dataNew": "dataNew", "data_new": "dataNew", "errMsg": "errMsg", "err_msg": "errMsg",
		"errNumber": "errNumber", "err_number": "errNumber",
		"createBy": "createBy", "createdBy": "createBy", "created_by": "createBy", "create_by": "createBy",
		"updateBy": "updateBy", "updatedBy": "updateBy", "updated_by": "updateBy", "update_by": "updateBy",
		"createdAt": "createdAt", "created_at": "createdAt", "updatedAt": "updatedAt", "updated_at": "updatedAt",
		"deletedAt": "deletedAt", "deleted_at": "deletedAt",
	}
}

func isResourceListControlField(name string) bool {
	switch name {
	case "", "pageIndex", "pageNum", "page", "current", "pageSize", "limit", "keyword", "search", "orderBy", "sort", "order", "sortOrder":
		return true
	default:
		return strings.HasSuffix(name, "Order")
	}
}

func isResourceListFilterAlreadyApplied(name string) bool {
	switch name {
	case "accountId", "appId", "groupId", "containerId", "unitId", "status", "passLevel":
		return true
	default:
		return false
	}
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
	addLegacyResourceResponseAliases(record, def)
	return record
}

func addLegacyResourceResponseAliases(record Record, def *resourceDefinition) {
	if record == nil || def == nil {
		return
	}
	if value, ok := record["createBy"]; ok {
		record["createBy"] = value
	}
	if value, ok := record["updateBy"]; ok {
		record["updateBy"] = value
	}
	switch def.name {
	case "account-details":
		copyRecordAlias(record, "grade", "nj")
		copyRecordAlias(record, "college", "xymc")
		copyRecordAlias(record, "collegeCode", "xydm")
		copyRecordAlias(record, "campus", "xq")
		copyRecordAlias(record, "schoolSystem", "xz")
		copyRecordAlias(record, "graduatedAt", "yjbysj")
		copyRecordAlias(record, "major", "zymc")
		copyRecordAlias(record, "className", "bjmc")
	case "pass-rules":
		copyRecordAlias(record, "intervalDays", "interval")
	}
}

func copyRecordAlias(record Record, source string, alias string) {
	if value, ok := record[source]; ok {
		record[alias] = value
	}
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

func (s *serviceImpl) actorID(ctx context.Context) int64 {
	tenantCtx := s.tenantFilter.Context(ctx)
	if tenantCtx.ActingUserID > 0 {
		return int64(tenantCtx.ActingUserID)
	}
	return int64(tenantCtx.UserID)
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

func normalizeLegacyResourceBody(resource string, body map[string]any) map[string]any {
	if body == nil {
		return map[string]any{}
	}
	setIfMissing := func(target string, aliases ...string) {
		if hasField(body, target) {
			return
		}
		for _, alias := range aliases {
			if hasField(body, alias) {
				body[target] = body[alias]
				return
			}
		}
	}
	switch strings.TrimSpace(resource) {
	case "accounts":
		setIfMissing("unitId", "unit_id", "groupId", "group_id")
	case "account-details", "account-details-legacy":
		setIfMissing("accountId", "account_id")
		setIfMissing("idcard", "idCard")
		setIfMissing("grade", "nj")
		setIfMissing("college", "xymc")
		setIfMissing("collegeCode", "college_code", "xydm")
		setIfMissing("campus", "xq")
		setIfMissing("schoolSystem", "school_system", "xz")
		setIfMissing("graduatedAt", "graduated_at", "yjbysj")
		setIfMissing("major", "zymc")
		setIfMissing("className", "class_name", "bjmc")
	case "pass-rules", "pass-ruler":
		setIfMissing("intervalDays", "interval", "interval_days")
		setIfMissing("intervalStatus", "interval_status")
	}
	return body
}

func commonTimeFields() map[string]struct{} {
	return map[string]struct{}{
		"effectAt":  {},
		"expireAt":  {},
		"createdAt": {},
		"updatedAt": {},
		"deletedAt": {},
		"loginTime": {},
		"startAt":   {},
		"endAt":     {},
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
			"id": cols.Id, "number": cols.Number, "name": cols.Name, "phone": cols.Phone,
			"effectAt": cols.EffectAt, "expireAt": cols.ExpireAt,
			"passLevel": cols.PassLevel, "containerId": cols.ContainerId, "unitId": cols.UnitId, "status": cols.Status,
			"createBy": cols.CreateBy, "updateBy": cols.UpdateBy, "createdAt": cols.CreatedAt, "updatedAt": cols.UpdatedAt, "deletedAt": cols.DeletedAt,
		},
		likeFields: map[string]struct{}{"name": {}, "phone": {}},
		timeFields: commonTimeFields(),
		model:      func(ctx context.Context) *gdb.Model { return dao.Account.Ctx(ctx) },
		data:       s.accountData,
	}
}

func (s *serviceImpl) accountDetailResource() *resourceDefinition {
	cols := dao.AccountDetails.Columns()
	return &resourceDefinition{
		name:          "account-details",
		table:         dao.AccountDetails.Table(),
		idColumn:      cols.AccountId,
		defaultOrder:  cols.AccountId,
		keywordFields: []string{cols.Email, cols.Wechat, cols.Idcard},
		apiToColumn: map[string]string{
			"accountId": cols.AccountId, "birthday": cols.Birthday, "email": cols.Email, "gender": cols.Gender,
			"qq": cols.Qq, "wechat": cols.Wechat, "idcard": cols.Idcard, "avatar": cols.Avatar, "source": cols.Source,
			"grade": cols.Nj, "college": cols.Xymc, "collegeCode": cols.Xydm, "campus": cols.Xq,
			"schoolSystem": cols.Xz, "graduatedAt": cols.Yjbysj, "major": cols.Zymc, "className": cols.Bjmc,
			"face": cols.Face, "createBy": cols.CreateBy, "updateBy": cols.UpdateBy, "createdAt": cols.CreatedAt, "updatedAt": cols.UpdatedAt,
			"nj": cols.Nj, "xymc": cols.Xymc, "xydm": cols.Xydm, "xq": cols.Xq, "xz": cols.Xz,
			"yjbysj": cols.Yjbysj, "zymc": cols.Zymc, "bjmc": cols.Bjmc,
		},
		timeFields: commonTimeFields(),
		dateFields: map[string]struct{}{"birthday": {}},
		model:      func(ctx context.Context) *gdb.Model { return dao.AccountDetails.Ctx(ctx) },
		data:       s.accountDetailData,
	}
}

func (s *serviceImpl) groupResource() *resourceDefinition {
	cols := dao.Groups.Columns()
	return simpleNamedResource("groups", dao.Groups.Table(), cols.Id, cols.Name, cols.Alias, cols.CreateBy, cols.UpdateBy, cols.CreatedAt, cols.UpdatedAt, cols.DeletedAt, func(ctx context.Context) *gdb.Model { return dao.Groups.Ctx(ctx) }, s.groupData)
}

func (s *serviceImpl) unitResource() *resourceDefinition {
	cols := dao.Units.Columns()
	def := simpleNamedResource("units", dao.Units.Table(), cols.Id, cols.Name, cols.Alias, cols.CreateBy, cols.UpdateBy, cols.CreatedAt, cols.UpdatedAt, cols.DeletedAt, func(ctx context.Context) *gdb.Model { return dao.Units.Ctx(ctx) }, s.unitData)
	def.apiToColumn["code"] = cols.Code
	def.keywordFields = append(def.keywordFields, cols.Code)
	return def
}

func (s *serviceImpl) containerResource() *resourceDefinition {
	cols := dao.Containers.Columns()
	def := simpleNamedResource("containers", dao.Containers.Table(), cols.Id, cols.Name, cols.Alias, cols.CreateBy, cols.UpdateBy, cols.CreatedAt, cols.UpdatedAt, cols.DeletedAt, func(ctx context.Context) *gdb.Model { return dao.Containers.Ctx(ctx) }, s.containerData)
	def.apiToColumn["accountCount"] = cols.AccountCount
	def.apiToColumn["adminCount"] = cols.AdminCount
	return def
}

func simpleNamedResource(name, table, id, resourceName, alias, createBy, updateBy, createdAt, updatedAt, deletedAt string, model func(context.Context) *gdb.Model, data func(context.Context, map[string]any, bool) (any, error)) *resourceDefinition {
	return &resourceDefinition{
		name:          name,
		table:         table,
		idColumn:      id,
		defaultOrder:  id,
		keywordFields: []string{resourceName, alias},
		apiToColumn: map[string]string{
			"id": id, "name": resourceName, "alias": alias, "createBy": createBy, "updateBy": updateBy,
			"createdAt": createdAt, "updatedAt": updatedAt, "deletedAt": deletedAt,
		},
		timeFields: commonTimeFields(),
		model:      model,
		data:       data,
	}
}

// The remaining resource definitions are placed in uidentity_resource_defs.go
// to keep this registry file readable.
