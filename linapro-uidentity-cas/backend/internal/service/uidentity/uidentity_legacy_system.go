// This file implements old go-admin system management compatibility backed by
// plugin-prefixed sys_* tables. It keeps legacy JSON field names while using
// generated DAO/DO objects and set-based reads.

package uidentity

import (
	"context"
	"strings"
	"time"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/util/gconv"
	"golang.org/x/crypto/bcrypt"

	"lina-core/pkg/bizerr"
	"lina-plugin-linapro-uidentity-cas/backend/internal/dao"
	"lina-plugin-linapro-uidentity-cas/backend/internal/model/do"
)

const (
	legacyDefaultAvatar       = "https://wpimg.wallstcn.com/f778738c-e4f8-4870-b634-56703b4acafe.gif"
	legacyAllPermission       = "*:*:*"
	legacySystemAdminRoleKey  = "admin"
	legacySystemAdminRoleName = "系统管理员"
)

type legacySystemResourceDefinition struct {
	resourceDefinition
}

type legacyTreeSpec struct {
	idField     string
	parentField string
	labelField  string
	childField  string
}

// ListLegacySystemResource returns old admin system resource rows from plugin tables.
func (s *serviceImpl) ListLegacySystemResource(ctx context.Context, in LegacySystemResourceListInput) (*ResourceListOutput, error) {
	def, err := s.legacySystemDefinition(in.Resource)
	if err != nil {
		return nil, err
	}
	model := s.applyLegacySystemFilters(ctx, def.model(ctx), def, in.Filters)
	total, err := model.Count()
	if err != nil {
		return nil, err
	}
	model = orderLegacySystemModel(model, def, in.OrderBy, in.Order)
	rows, err := model.
		Fields(projectionFields(&def.resourceDefinition)...).
		Page(in.PageNum, in.PageSize).
		All()
	if err != nil {
		return nil, err
	}
	return &ResourceListOutput{List: projectResult(rows, &def.resourceDefinition), Total: total}, nil
}

// GetLegacySystemResource returns one old admin system resource row.
func (s *serviceImpl) GetLegacySystemResource(ctx context.Context, resource string, id int64) (Record, error) {
	def, err := s.legacySystemDefinition(resource)
	if err != nil {
		return nil, err
	}
	row, err := def.model(ctx).
		Fields(projectionFields(&def.resourceDefinition)...).
		Where(def.idColumn, id).
		One()
	if err != nil {
		return nil, err
	}
	if row.IsEmpty() {
		return nil, bizerr.NewCode(CodeResourceNotFound)
	}
	return projectRecord(row, &def.resourceDefinition), nil
}

// CreateLegacySystemResource inserts one old admin system resource row.
func (s *serviceImpl) CreateLegacySystemResource(ctx context.Context, resource string, body map[string]any) (int64, error) {
	def, err := s.legacySystemDefinition(resource)
	if err != nil {
		return 0, err
	}
	data, err := def.data(ctx, normalizeLegacySystemBody(resource, body), true)
	if err != nil {
		return 0, err
	}
	return def.model(ctx).Data(data).InsertAndGetId()
}

// UpdateLegacySystemResource updates one old admin system resource row.
func (s *serviceImpl) UpdateLegacySystemResource(ctx context.Context, resource string, id int64, body map[string]any) error {
	def, err := s.legacySystemDefinition(resource)
	if err != nil {
		return err
	}
	if _, err := s.GetLegacySystemResource(ctx, resource, id); err != nil {
		return err
	}
	data, err := def.data(ctx, normalizeLegacySystemBody(resource, body), false)
	if err != nil {
		return err
	}
	_, err = def.model(ctx).
		Where(def.idColumn, id).
		OmitNilData().
		Data(data).
		Update()
	return err
}

// DeleteLegacySystemResource deletes old admin system resource rows.
func (s *serviceImpl) DeleteLegacySystemResource(ctx context.Context, resource string, ids string) error {
	def, err := s.legacySystemDefinition(resource)
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
	count, err := def.model(ctx).WhereIn(def.idColumn, idList).Count()
	if err != nil {
		return err
	}
	if count != len(idList) {
		return bizerr.NewCode(CodeResourceNotFound)
	}
	_, err = def.model(ctx).WhereIn(def.idColumn, idList).Delete()
	return err
}

// LegacyDeptTree returns old /dept full-record tree rows.
func (s *serviceImpl) LegacyDeptTree(ctx context.Context, filters map[string]any) ([]Record, error) {
	def, err := s.legacySystemDefinition("depts")
	if err != nil {
		return nil, err
	}
	cols := dao.SysDept.Columns()
	rows, err := s.applyLegacySystemFilters(ctx, def.model(ctx), def, filters).
		Fields(projectionFields(&def.resourceDefinition)...).
		OrderAsc(cols.Sort).
		OrderAsc(cols.DeptId).
		All()
	if err != nil {
		return nil, err
	}
	records := projectResult(rows, &def.resourceDefinition)
	return buildLegacyRecordTree(records, legacyTreeSpec{"deptId", "parentId", "", "children"}), nil
}

// LegacyDeptLabelTree returns old /deptTree label tree rows.
func (s *serviceImpl) LegacyDeptLabelTree(ctx context.Context, filters map[string]any) ([]Record, error) {
	records, err := s.LegacyDeptTree(ctx, filters)
	if err != nil {
		return nil, err
	}
	return legacyLabelTreeFromRecords(records, "deptId", "deptName"), nil
}

// LegacyRoleDeptTreeSelect returns depts and checkedKeys for old role dept selection.
func (s *serviceImpl) LegacyRoleDeptTreeSelect(ctx context.Context, roleID int64) (Record, error) {
	tree, err := s.LegacyDeptLabelTree(ctx, nil)
	if err != nil {
		return nil, err
	}
	checked, err := s.legacyRoleDeptIDs(ctx, roleID)
	if err != nil {
		return nil, err
	}
	return Record{"depts": tree, "checkedKeys": checked}, nil
}

// LegacyMenuRoleTree returns old menu role tree rows.
func (s *serviceImpl) LegacyMenuRoleTree(ctx context.Context, roleID int64) ([]Record, error) {
	def, err := s.legacySystemDefinition("menus")
	if err != nil {
		return nil, err
	}
	cols := dao.SysMenu.Columns()
	model := def.model(ctx).
		Fields(projectionFields(&def.resourceDefinition)...).
		OrderAsc(cols.Sort).
		OrderAsc(cols.MenuId)
	if roleID > 0 {
		menuIDs, err := s.legacyRoleMenuIDs(ctx, roleID)
		if err != nil {
			return nil, err
		}
		if len(menuIDs) == 0 {
			return []Record{}, nil
		}
		model = model.WhereIn(cols.MenuId, menuIDs)
	}
	rows, err := model.All()
	if err != nil {
		return nil, err
	}
	records := projectResult(rows, &def.resourceDefinition)
	return buildLegacyRecordTree(records, legacyTreeSpec{"menuId", "parentId", "", "children"}), nil
}

// LegacyRoleMenuTreeSelect returns menus and checkedKeys for old role menu selection.
func (s *serviceImpl) LegacyRoleMenuTreeSelect(ctx context.Context, roleID int64) (Record, error) {
	tree, err := s.legacyMenuLabelTree(ctx)
	if err != nil {
		return nil, err
	}
	checked, err := s.legacyRoleMenuIDs(ctx, roleID)
	if err != nil {
		return nil, err
	}
	return Record{"menus": tree, "checkedKeys": checked}, nil
}

// LegacyDictTypeOptions returns old dictionary type option rows.
func (s *serviceImpl) LegacyDictTypeOptions(ctx context.Context) ([]Record, error) {
	def, err := s.legacySystemDefinition("dict-types")
	if err != nil {
		return nil, err
	}
	cols := dao.SysDictType.Columns()
	rows, err := def.model(ctx).
		Fields(projectionFields(&def.resourceDefinition)...).
		OrderAsc(cols.DictId).
		All()
	if err != nil {
		return nil, err
	}
	return projectResult(rows, &def.resourceDefinition), nil
}

// LegacyDictDataOptions returns old dictionary data label/value options.
func (s *serviceImpl) LegacyDictDataOptions(ctx context.Context, dictType string) ([]Record, error) {
	cols := dao.SysDictData.Columns()
	model := dao.SysDictData.Ctx(ctx).Fields(cols.DictLabel, cols.DictValue).OrderAsc(cols.DictSort).OrderAsc(cols.DictCode)
	if strings.TrimSpace(dictType) != "" {
		model = model.Where(cols.DictType, strings.TrimSpace(dictType))
	}
	rows, err := model.All()
	if err != nil {
		return nil, err
	}
	result := make([]Record, 0, len(rows))
	for _, row := range rows {
		result = append(result, Record{"label": row[cols.DictLabel].Interface(), "value": row[cols.DictValue].Interface()})
	}
	return result, nil
}

// LegacyConfigByKey returns a configKey/configValue payload.
func (s *serviceImpl) LegacyConfigByKey(ctx context.Context, configKey string) (Record, error) {
	cols := dao.SysConfig.Columns()
	row, err := dao.SysConfig.Ctx(ctx).
		Fields(cols.ConfigKey, cols.ConfigValue).
		Where(cols.ConfigKey, strings.TrimSpace(configKey)).
		One()
	if err != nil {
		return nil, err
	}
	if row.IsEmpty() {
		return nil, bizerr.NewCode(CodeResourceNotFound)
	}
	return Record{"configKey": row[cols.ConfigKey].String(), "configValue": row[cols.ConfigValue].String()}, nil
}

// LegacyFrontendConfigs returns old app-config key/value rows.
func (s *serviceImpl) LegacyFrontendConfigs(ctx context.Context) (Record, error) {
	cols := dao.SysConfig.Columns()
	rows, err := dao.SysConfig.Ctx(ctx).
		Fields(cols.ConfigKey, cols.ConfigValue).
		Where(cols.IsFrontend, "1").
		OrderAsc(cols.Id).
		All()
	if err != nil {
		return nil, err
	}
	return legacyConfigRowsToMap(rows, cols.ConfigKey, cols.ConfigValue), nil
}

// LegacySetConfigs returns old set-config key/value rows.
func (s *serviceImpl) LegacySetConfigs(ctx context.Context) (Record, error) {
	cols := dao.SysConfig.Columns()
	rows, err := dao.SysConfig.Ctx(ctx).
		Fields(cols.ConfigKey, cols.ConfigValue).
		OrderAsc(cols.Id).
		All()
	if err != nil {
		return nil, err
	}
	return legacyConfigRowsToMap(rows, cols.ConfigKey, cols.ConfigValue), nil
}

// UpdateLegacySetConfigs updates existing config keys with supplied values.
func (s *serviceImpl) UpdateLegacySetConfigs(ctx context.Context, values map[string]string) error {
	cols := dao.SysConfig.Columns()
	for key, value := range values {
		trimmedKey := strings.TrimSpace(key)
		if trimmedKey == "" {
			continue
		}
		_, err := dao.SysConfig.Ctx(ctx).
			Where(cols.ConfigKey, trimmedKey).
			OmitNilData().
			Data(do.SysConfig{ConfigValue: value, UpdateBy: s.actorID(ctx)}).
			Update()
		if err != nil {
			return err
		}
	}
	return nil
}

// LegacyGetInfo returns old /getinfo metadata for the current actor.
func (s *serviceImpl) LegacyGetInfo(ctx context.Context) (Record, error) {
	userID := s.actorID(ctx)
	if userID <= 0 {
		return nil, bizerr.NewCode(CodeResourceNotFound)
	}
	user, err := s.GetLegacySystemResource(ctx, "sys-users", userID)
	if err != nil {
		return nil, err
	}
	roleID := gconv.Int64(user["roleId"])
	role, _ := s.legacyRoleRecord(ctx, roleID)
	roleName := gconv.String(role["roleName"])
	roleKey := gconv.String(role["roleKey"])
	roles := []string{roleName}
	if strings.TrimSpace(roleName) == "" {
		roles = []string{roleKey}
	}
	permissions := []string{legacyAllPermission}
	if !isLegacyAdminRole(roleName, roleKey) && roleID > 0 {
		loaded, err := s.legacyRolePermissions(ctx, roleID)
		if err != nil {
			return nil, err
		}
		if len(loaded) > 0 {
			permissions = loaded
		}
	}
	avatar := gconv.String(user["avatar"])
	if strings.TrimSpace(avatar) == "" {
		avatar = legacyDefaultAvatar
	}
	name := gconv.String(user["nickName"])
	return Record{
		"roles":        roles,
		"permissions":  permissions,
		"buttons":      permissions,
		"introduction": " am a super administrator",
		"avatar":       avatar,
		"userName":     name,
		"userId":       user["userId"],
		"deptId":       user["deptId"],
		"name":         name,
		"code":         200,
	}, nil
}

// UpdateLegacySysUserAvatar updates old sys_user.avatar.
func (s *serviceImpl) UpdateLegacySysUserAvatar(ctx context.Context, userID int64, avatar string) error {
	if userID <= 0 {
		userID = s.actorID(ctx)
	}
	if userID <= 0 || strings.TrimSpace(avatar) == "" {
		return bizerr.NewCode(CodeResourceNotFound)
	}
	cols := dao.SysUser.Columns()
	result, err := dao.SysUser.Ctx(ctx).
		Where(cols.UserId, userID).
		OmitNilData().
		Data(do.SysUser{Avatar: strings.TrimSpace(avatar), UpdateBy: s.actorID(ctx)}).
		Update()
	if err != nil {
		return err
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return bizerr.NewCode(CodeResourceNotFound)
	}
	return nil
}

// UpdateLegacySysUserPassword updates old sys_user.password.
func (s *serviceImpl) UpdateLegacySysUserPassword(ctx context.Context, userID int64, oldPassword string, newPassword string, requireOld bool) error {
	if userID <= 0 {
		userID = s.actorID(ctx)
	}
	if userID <= 0 || strings.TrimSpace(newPassword) == "" {
		return bizerr.NewCode(CodeResourceNotFound)
	}
	cols := dao.SysUser.Columns()
	if requireOld {
		row, err := dao.SysUser.Ctx(ctx).Fields(cols.Password).Where(cols.UserId, userID).One()
		if err != nil {
			return err
		}
		if row.IsEmpty() {
			return bizerr.NewCode(CodeResourceNotFound)
		}
		if err := bcrypt.CompareHashAndPassword([]byte(row[cols.Password].String()), []byte(oldPassword)); err != nil {
			return bizerr.WrapCode(err, CodeInvalidCredentials)
		}
	}
	result, err := dao.SysUser.Ctx(ctx).
		Where(cols.UserId, userID).
		OmitNilData().
		Data(do.SysUser{Password: strings.TrimSpace(newPassword), UpdateBy: s.actorID(ctx)}).
		Update()
	if err != nil {
		return err
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return bizerr.NewCode(CodeResourceNotFound)
	}
	return nil
}

// UpdateLegacySysUserStatus updates old sys_user.status.
func (s *serviceImpl) UpdateLegacySysUserStatus(ctx context.Context, userID int64, status string) error {
	if userID <= 0 {
		userID = s.actorID(ctx)
	}
	if userID <= 0 || strings.TrimSpace(status) == "" {
		return bizerr.NewCode(CodeResourceNotFound)
	}
	cols := dao.SysUser.Columns()
	result, err := dao.SysUser.Ctx(ctx).
		Where(cols.UserId, userID).
		OmitNilData().
		Data(do.SysUser{Status: strings.TrimSpace(status), UpdateBy: s.actorID(ctx)}).
		Update()
	if err != nil {
		return err
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return bizerr.NewCode(CodeResourceNotFound)
	}
	return nil
}

func (s *serviceImpl) legacySystemDefinition(resource string) (*legacySystemResourceDefinition, error) {
	defs := s.legacySystemDefinitions()
	if def, ok := defs[strings.TrimSpace(resource)]; ok {
		return def, nil
	}
	return nil, bizerr.NewCode(CodeResourceNotSupported)
}

func (s *serviceImpl) legacySystemDefinitions() map[string]*legacySystemResourceDefinition {
	return map[string]*legacySystemResourceDefinition{
		"sys-users":      s.legacySysUserResource(),
		"sys-user":       s.legacySysUserResource(),
		"login-logs":     s.legacySysLoginLogResource(),
		"opera-logs":     s.legacySysOperaLogResource(),
		"apis":           s.legacySysAPIResource(),
		"depts":          s.legacySysDeptResource(),
		"posts":          s.legacySysPostResource(),
		"dict-types":     s.legacySysDictTypeResource(),
		"dict-data":      s.legacySysDictDataResource(),
		"configs":        s.legacySysConfigResource(),
		"roles":          s.legacySysRoleResource(),
		"menus":          s.legacySysMenuResource(),
		"sys-tables":     s.legacySysTablesResource(),
		"sys-columns":    s.legacySysColumnsResource(),
		"db-tables":      s.legacySysTablesResource(),
		"db-columns":     s.legacySysColumnsResource(),
		"sys-table-info": s.legacySysTablesResource(),
	}
}

func (s *serviceImpl) legacySysUserResource() *legacySystemResourceDefinition {
	cols := dao.SysUser.Columns()
	return legacySystemResource("sys-users", cols.UserId, cols.UserId, []string{cols.Username, cols.NickName, cols.Phone}, map[string]string{
		"userId": cols.UserId, "username": cols.Username, "nickName": cols.NickName,
		"phone": cols.Phone, "roleId": cols.RoleId, "avatar": cols.Avatar, "sex": cols.Sex, "email": cols.Email,
		"deptId": cols.DeptId, "postId": cols.PostId, "remark": cols.Remark, "status": cols.Status,
		"createBy": cols.CreateBy, "updateBy": cols.UpdateBy, "createdAt": cols.CreatedAt, "updatedAt": cols.UpdatedAt, "deletedAt": cols.DeletedAt,
	}, map[string]struct{}{"username": {}, "nickName": {}, "phone": {}, "email": {}}, func(ctx context.Context) *gdb.Model { return dao.SysUser.Ctx(ctx) }, s.legacySysUserData)
}

func (s *serviceImpl) legacySysLoginLogResource() *legacySystemResourceDefinition {
	cols := dao.SysLoginLog.Columns()
	return legacySystemResource("login-logs", cols.Id, cols.CreatedAt, []string{cols.Username, cols.Ipaddr, cols.LoginLocation}, map[string]string{
		"id": cols.Id, "username": cols.Username, "status": cols.Status, "ipaddr": cols.Ipaddr, "loginLocation": cols.LoginLocation,
		"browser": cols.Browser, "os": cols.Os, "platform": cols.Platform, "loginTime": cols.LoginTime,
		"remark": cols.Remark, "msg": cols.Msg, "createdAt": cols.CreatedAt, "updatedAt": cols.UpdatedAt, "createBy": cols.CreateBy, "updateBy": cols.UpdateBy,
	}, map[string]struct{}{"username": {}, "ipaddr": {}, "loginLocation": {}}, func(ctx context.Context) *gdb.Model { return dao.SysLoginLog.Ctx(ctx) }, s.legacySysLoginLogData)
}

func (s *serviceImpl) legacySysOperaLogResource() *legacySystemResourceDefinition {
	cols := dao.SysOperaLog.Columns()
	return legacySystemResource("opera-logs", cols.Id, cols.CreatedAt, []string{cols.Title, cols.Method, cols.OperUrl, cols.OperIp}, map[string]string{
		"id": cols.Id, "title": cols.Title, "businessType": cols.BusinessType, "businessTypes": cols.BusinessTypes,
		"method": cols.Method, "requestMethod": cols.RequestMethod, "operatorType": cols.OperatorType, "operName": cols.OperName,
		"deptName": cols.DeptName, "operUrl": cols.OperUrl, "operIp": cols.OperIp, "operLocation": cols.OperLocation,
		"operParam": cols.OperParam, "status": cols.Status, "operTime": cols.OperTime, "jsonResult": cols.JsonResult,
		"remark": cols.Remark, "latencyTime": cols.LatencyTime, "userAgent": cols.UserAgent, "createdAt": cols.CreatedAt,
		"updatedAt": cols.UpdatedAt, "createBy": cols.CreateBy, "updateBy": cols.UpdateBy,
	}, map[string]struct{}{"title": {}, "method": {}, "requestMethod": {}, "operUrl": {}}, func(ctx context.Context) *gdb.Model { return dao.SysOperaLog.Ctx(ctx) }, s.legacySysOperaLogData)
}

func (s *serviceImpl) legacySysAPIResource() *legacySystemResourceDefinition {
	cols := dao.SysApi.Columns()
	return legacySystemResource("apis", cols.Id, cols.Id, []string{cols.Title, cols.Path}, map[string]string{
		"id": cols.Id, "handle": cols.Handle, "title": cols.Title, "path": cols.Path, "type": cols.Type, "action": cols.Action,
		"createBy": cols.CreateBy, "updateBy": cols.UpdateBy, "createdAt": cols.CreatedAt, "updatedAt": cols.UpdatedAt, "deletedAt": cols.DeletedAt,
	}, map[string]struct{}{"title": {}, "path": {}}, func(ctx context.Context) *gdb.Model { return dao.SysApi.Ctx(ctx) }, s.legacySysAPIData)
}

func (s *serviceImpl) legacySysDeptResource() *legacySystemResourceDefinition {
	cols := dao.SysDept.Columns()
	return legacySystemResource("depts", cols.DeptId, cols.Sort, []string{cols.DeptName, cols.Phone, cols.Email}, map[string]string{
		"deptId": cols.DeptId, "parentId": cols.ParentId, "deptPath": cols.DeptPath, "deptName": cols.DeptName,
		"sort": cols.Sort, "leader": cols.Leader, "phone": cols.Phone, "email": cols.Email, "status": cols.Status,
		"createBy": cols.CreateBy, "updateBy": cols.UpdateBy, "createdAt": cols.CreatedAt, "updatedAt": cols.UpdatedAt, "deletedAt": cols.DeletedAt,
	}, map[string]struct{}{"deptName": {}, "leader": {}, "phone": {}, "email": {}}, func(ctx context.Context) *gdb.Model { return dao.SysDept.Ctx(ctx) }, s.legacySysDeptData)
}

func (s *serviceImpl) legacySysPostResource() *legacySystemResourceDefinition {
	cols := dao.SysPost.Columns()
	return legacySystemResource("posts", cols.PostId, cols.Sort, []string{cols.PostName, cols.PostCode}, map[string]string{
		"postId": cols.PostId, "postName": cols.PostName, "postCode": cols.PostCode, "sort": cols.Sort, "status": cols.Status,
		"remark": cols.Remark, "createBy": cols.CreateBy, "updateBy": cols.UpdateBy, "createdAt": cols.CreatedAt, "updatedAt": cols.UpdatedAt, "deletedAt": cols.DeletedAt,
	}, map[string]struct{}{"postName": {}, "postCode": {}}, func(ctx context.Context) *gdb.Model { return dao.SysPost.Ctx(ctx) }, s.legacySysPostData)
}

func (s *serviceImpl) legacySysDictTypeResource() *legacySystemResourceDefinition {
	cols := dao.SysDictType.Columns()
	return legacySystemResource("dict-types", cols.DictId, cols.DictId, []string{cols.DictName, cols.DictType}, map[string]string{
		"id": cols.DictId, "dictId": cols.DictId, "dictName": cols.DictName, "dictType": cols.DictType, "status": cols.Status,
		"remark": cols.Remark, "createBy": cols.CreateBy, "updateBy": cols.UpdateBy, "createdAt": cols.CreatedAt, "updatedAt": cols.UpdatedAt, "deletedAt": cols.DeletedAt,
	}, map[string]struct{}{"dictName": {}, "dictType": {}}, func(ctx context.Context) *gdb.Model { return dao.SysDictType.Ctx(ctx) }, s.legacySysDictTypeData)
}

func (s *serviceImpl) legacySysDictDataResource() *legacySystemResourceDefinition {
	cols := dao.SysDictData.Columns()
	return legacySystemResource("dict-data", cols.DictCode, cols.DictSort, []string{cols.DictLabel, cols.DictValue, cols.DictType}, map[string]string{
		"id": cols.DictCode, "dictCode": cols.DictCode, "dictSort": cols.DictSort, "dictLabel": cols.DictLabel,
		"dictValue": cols.DictValue, "dictType": cols.DictType, "cssClass": cols.CssClass, "listClass": cols.ListClass,
		"isDefault": cols.IsDefault, "status": cols.Status, "default": cols.Default, "remark": cols.Remark,
		"createBy": cols.CreateBy, "updateBy": cols.UpdateBy, "createdAt": cols.CreatedAt, "updatedAt": cols.UpdatedAt, "deletedAt": cols.DeletedAt,
	}, map[string]struct{}{"dictLabel": {}, "dictValue": {}, "dictType": {}}, func(ctx context.Context) *gdb.Model { return dao.SysDictData.Ctx(ctx) }, s.legacySysDictDataData)
}

func (s *serviceImpl) legacySysConfigResource() *legacySystemResourceDefinition {
	cols := dao.SysConfig.Columns()
	return legacySystemResource("configs", cols.Id, cols.Id, []string{cols.ConfigName, cols.ConfigKey}, map[string]string{
		"id": cols.Id, "configName": cols.ConfigName, "configKey": cols.ConfigKey, "configValue": cols.ConfigValue,
		"configType": cols.ConfigType, "isFrontend": cols.IsFrontend, "remark": cols.Remark,
		"createBy": cols.CreateBy, "updateBy": cols.UpdateBy, "createdAt": cols.CreatedAt, "updatedAt": cols.UpdatedAt, "deletedAt": cols.DeletedAt,
	}, map[string]struct{}{"configName": {}, "configKey": {}}, func(ctx context.Context) *gdb.Model { return dao.SysConfig.Ctx(ctx) }, s.legacySysConfigData)
}

func (s *serviceImpl) legacySysRoleResource() *legacySystemResourceDefinition {
	cols := dao.SysRole.Columns()
	return legacySystemResource("roles", cols.RoleId, cols.RoleSort, []string{cols.RoleName, cols.RoleKey}, map[string]string{
		"roleId": cols.RoleId, "roleName": cols.RoleName, "status": cols.Status, "roleKey": cols.RoleKey, "roleSort": cols.RoleSort,
		"flag": cols.Flag, "remark": cols.Remark, "admin": cols.Admin, "dataScope": cols.DataScope,
		"createBy": cols.CreateBy, "updateBy": cols.UpdateBy, "createdAt": cols.CreatedAt, "updatedAt": cols.UpdatedAt, "deletedAt": cols.DeletedAt,
	}, map[string]struct{}{"roleName": {}, "roleKey": {}}, func(ctx context.Context) *gdb.Model { return dao.SysRole.Ctx(ctx) }, s.legacySysRoleData)
}

func (s *serviceImpl) legacySysMenuResource() *legacySystemResourceDefinition {
	cols := dao.SysMenu.Columns()
	return legacySystemResource("menus", cols.MenuId, cols.Sort, []string{cols.Title, cols.Path, cols.Permission}, map[string]string{
		"menuId": cols.MenuId, "menuName": cols.MenuName, "title": cols.Title, "icon": cols.Icon, "path": cols.Path,
		"paths": cols.Paths, "menuType": cols.MenuType, "action": cols.Action, "permission": cols.Permission,
		"parentId": cols.ParentId, "noCache": cols.NoCache, "breadcrumb": cols.Breadcrumb, "component": cols.Component,
		"sort": cols.Sort, "visible": cols.Visible, "isFrame": cols.IsFrame,
		"createBy": cols.CreateBy, "updateBy": cols.UpdateBy, "createdAt": cols.CreatedAt, "updatedAt": cols.UpdatedAt, "deletedAt": cols.DeletedAt,
	}, map[string]struct{}{"title": {}, "path": {}, "permission": {}}, func(ctx context.Context) *gdb.Model { return dao.SysMenu.Ctx(ctx) }, s.legacySysMenuData)
}

func (s *serviceImpl) legacySysTablesResource() *legacySystemResourceDefinition {
	cols := dao.SysTables.Columns()
	return legacySystemResource("sys-tables", cols.TableId, cols.TableId, []string{cols.TableName, cols.TableComment, cols.ClassName}, map[string]string{
		"tableId": cols.TableId, "tableName": cols.TableName, "tableComment": cols.TableComment, "className": cols.ClassName,
		"tplCategory": cols.TplCategory, "packageName": cols.PackageName, "moduleName": cols.ModuleName, "moduleFrontName": cols.ModuleFrontName,
		"businessName": cols.BusinessName, "functionName": cols.FunctionName, "functionAuthor": cols.FunctionAuthor, "pkColumn": cols.PkColumn,
		"pkGoField": cols.PkGoField, "pkJsonField": cols.PkJsonField, "options": cols.Options, "treeCode": cols.TreeCode,
		"treeParentCode": cols.TreeParentCode, "treeName": cols.TreeName, "tree": cols.Tree, "crud": cols.Crud, "remark": cols.Remark,
		"isDataScope": cols.IsDataScope, "isActions": cols.IsActions, "isAuth": cols.IsAuth, "isLogicalDelete": cols.IsLogicalDelete,
		"logicalDelete": cols.LogicalDelete, "logicalDeleteColumn": cols.LogicalDeleteColumn,
		"createBy": cols.CreateBy, "updateBy": cols.UpdateBy, "createdAt": cols.CreatedAt, "updatedAt": cols.UpdatedAt, "deletedAt": cols.DeletedAt,
	}, map[string]struct{}{"tableName": {}, "tableComment": {}, "className": {}}, func(ctx context.Context) *gdb.Model { return dao.SysTables.Ctx(ctx) }, s.legacySysTablesData)
}

func (s *serviceImpl) legacySysColumnsResource() *legacySystemResourceDefinition {
	cols := dao.SysColumns.Columns()
	return legacySystemResource("sys-columns", cols.ColumnId, cols.Sort, []string{cols.ColumnName, cols.ColumnComment, cols.GoField, cols.JsonField}, map[string]string{
		"columnId": cols.ColumnId, "tableId": cols.TableId, "columnName": cols.ColumnName, "columnComment": cols.ColumnComment,
		"columnType": cols.ColumnType, "goType": cols.GoType, "goField": cols.GoField, "jsonField": cols.JsonField,
		"isPk": cols.IsPk, "isIncrement": cols.IsIncrement, "isRequired": cols.IsRequired, "isInsert": cols.IsInsert,
		"isEdit": cols.IsEdit, "isList": cols.IsList, "isQuery": cols.IsQuery, "queryType": cols.QueryType,
		"htmlType": cols.HtmlType, "dictType": cols.DictType, "sort": cols.Sort, "list": cols.List, "pk": cols.Pk,
		"required": cols.Required, "superColumn": cols.SuperColumn, "usableColumn": cols.UsableColumn, "increment": cols.Increment,
		"insert": cols.Insert, "edit": cols.Edit, "query": cols.Query, "remark": cols.Remark,
		"fkTableName": cols.FkTableName, "fkTableNameClass": cols.FkTableNameClass, "fkTableNamePackage": cols.FkTableNamePackage,
		"fkLabelId": cols.FkLabelId, "fkLabelName": cols.FkLabelName,
		"createBy": cols.CreateBy, "updateBy": cols.UpdateBy, "createdAt": cols.CreatedAt, "updatedAt": cols.UpdatedAt, "deletedAt": cols.DeletedAt,
	}, map[string]struct{}{"columnName": {}, "columnComment": {}, "goField": {}, "jsonField": {}}, func(ctx context.Context) *gdb.Model { return dao.SysColumns.Ctx(ctx) }, s.legacySysColumnsData)
}

func legacySystemResource(name string, idColumn string, defaultOrder string, keywordFields []string, apiToColumn map[string]string, likeFields map[string]struct{}, model func(context.Context) *gdb.Model, data func(context.Context, map[string]any, bool) (any, error)) *legacySystemResourceDefinition {
	return &legacySystemResourceDefinition{resourceDefinition: resourceDefinition{
		name:          name,
		idColumn:      idColumn,
		defaultOrder:  defaultOrder,
		keywordFields: keywordFields,
		apiToColumn:   apiToColumn,
		likeFields:    likeFields,
		timeFields:    commonTimeFields(),
		model:         model,
		data:          data,
	}}
}

func (s *serviceImpl) applyLegacySystemFilters(ctx context.Context, model *gdb.Model, def *legacySystemResourceDefinition, filters map[string]any) *gdb.Model {
	if len(filters) == 0 {
		return model
	}
	keyword := strings.TrimSpace(gconv.String(filters["keyword"]))
	if keyword == "" {
		keyword = strings.TrimSpace(gconv.String(filters["search"]))
	}
	if keyword != "" {
		model = applyLegacyKeyword(model, def.keywordFields, keyword)
	}
	for fieldName, value := range filters {
		apiName := legacySystemFilterAPIName(fieldName)
		if isResourceListControlField(apiName) || apiName == "keyword" || apiName == "search" {
			continue
		}
		column := def.apiToColumn[apiName]
		if (apiName == "beginTime" || apiName == "endTime") && column == "" {
			column = legacySystemRangeColumn(def)
		}
		if column == "" || strings.TrimSpace(gconv.String(value)) == "" {
			continue
		}
		if def.name == "sys-users" && apiName == "deptId" {
			deptCols := dao.SysDept.Columns()
			deptIDs := dao.SysDept.Ctx(ctx).
				Fields(deptCols.DeptId).
				WhereLike(deptCols.DeptPath, "%"+strings.TrimSpace(gconv.String(value))+"%")
			model = model.Where(column+" IN (?)", deptIDs)
			continue
		}
		switch apiName {
		case "beginTime":
			model = model.WhereGTE(column, gconv.Time(value))
		case "endTime":
			model = model.WhereLTE(column, gconv.Time(value))
		default:
			if _, ok := def.likeFields[apiName]; ok {
				model = model.WhereLike(column, "%"+strings.TrimSpace(gconv.String(value))+"%")
			} else {
				model = model.Where(column, value)
			}
		}
	}
	return model
}

func legacySystemRangeColumn(def *legacySystemResourceDefinition) string {
	if def == nil {
		return ""
	}
	return def.apiToColumn["createdAt"]
}

func applyLegacyKeyword(model *gdb.Model, fields []string, keyword string) *gdb.Model {
	if len(fields) == 0 {
		return model
	}
	conditions := make([]string, 0, len(fields))
	values := make([]any, 0, len(fields))
	for _, field := range fields {
		conditions = append(conditions, field+" LIKE ?")
		values = append(values, "%"+keyword+"%")
	}
	return model.Where("("+strings.Join(conditions, " OR ")+")", values...)
}

func orderLegacySystemModel(model *gdb.Model, def *legacySystemResourceDefinition, orderBy string, order string) *gdb.Model {
	orderColumn := def.defaultOrder
	if apiName := legacySystemFilterAPIName(orderBy); apiName != "" {
		if column := def.apiToColumn[apiName]; column != "" {
			orderColumn = column
		}
	}
	if orderColumn == "" {
		return model
	}
	if strings.EqualFold(order, "asc") {
		return model.OrderAsc(orderColumn)
	}
	return model.OrderDesc(orderColumn)
}

func legacySystemFilterAPIName(name string) string {
	trimmed := strings.TrimSuffix(strings.TrimSpace(name), "Order")
	if trimmed == "" {
		return ""
	}
	if alias := resourceLegacyFieldAliases()[trimmed]; alias != "" {
		return alias
	}
	return legacySystemFieldAliases()[trimmed]
}

func legacySystemFieldAliases() map[string]string {
	return map[string]string{
		"ID": "id", "Id": "id",
		"handle": "handle", "title": "title", "path": "path", "type": "type", "action": "action",
		"sort": "sort", "status": "status", "remark": "remark", "phone": "phone", "email": "email", "leader": "leader",
		"userId": "userId", "user_id": "userId", "username": "username", "nickName": "nickName", "nick_name": "nickName",
		"roleId": "roleId", "role_id": "roleId", "deptId": "deptId", "dept_id": "deptId", "postId": "postId", "post_id": "postId",
		"parentId": "parentId", "parent_id": "parentId", "deptPath": "deptPath", "dept_path": "deptPath", "deptName": "deptName", "dept_name": "deptName",
		"postName": "postName", "post_name": "postName", "postCode": "postCode", "post_code": "postCode",
		"dictId": "dictId", "dict_id": "dictId", "dictCode": "dictCode", "dict_code": "dictCode", "dictName": "dictName", "dict_name": "dictName",
		"dictType": "dictType", "dict_type": "dictType", "dictSort": "dictSort", "dict_sort": "dictSort", "dictLabel": "dictLabel",
		"dict_label": "dictLabel", "dictValue": "dictValue", "dict_value": "dictValue", "cssClass": "cssClass", "css_class": "cssClass",
		"listClass": "listClass", "list_class": "listClass", "isDefault": "isDefault", "is_default": "isDefault",
		"configName": "configName", "config_name": "configName", "configKey": "configKey", "config_key": "configKey",
		"configType": "configType", "config_type": "configType", "configValue": "configValue", "config_value": "configValue",
		"isFrontend": "isFrontend", "is_frontend": "isFrontend",
		"menuId": "menuId", "menu_id": "menuId", "menuName": "menuName", "menu_name": "menuName", "menuType": "menuType", "menu_type": "menuType",
		"permission": "permission", "noCache": "noCache", "no_cache": "noCache", "isFrame": "isFrame", "is_frame": "isFrame",
		"requestMethod": "requestMethod", "request_method": "requestMethod", "businessType": "businessType", "business_type": "businessType",
		"businessTypes": "businessTypes", "business_types": "businessTypes", "operatorType": "operatorType", "operator_type": "operatorType",
		"operName": "operName", "oper_name": "operName", "operUrl": "operUrl",
		"oper_url": "operUrl", "operIp": "operIp", "oper_ip": "operIp", "operLocation": "operLocation", "oper_location": "operLocation",
		"operParam": "operParam", "oper_param": "operParam", "operTime": "operTime", "oper_time": "operTime",
		"jsonResult": "jsonResult", "json_result": "jsonResult", "latencyTime": "latencyTime", "latency_time": "latencyTime",
		"userAgent": "userAgent", "user_agent": "userAgent", "beginTime": "beginTime", "endTime": "endTime",
		"tableId": "tableId", "table_id": "tableId", "tableComment": "tableComment", "table_comment": "tableComment",
		"className": "className", "class_name": "className", "columnId": "columnId", "column_id": "columnId",
		"columnName": "columnName", "column_name": "columnName", "columnComment": "columnComment", "column_comment": "columnComment",
		"goField": "goField", "go_field": "goField", "jsonField": "jsonField", "json_field": "jsonField",
	}
}

func normalizeLegacySystemBody(resource string, body map[string]any) map[string]any {
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
	setIfMissing("id", "Id", "ID")
	setIfMissing("userId", "user_id", "id")
	setIfMissing("deptId", "dept_id", "id")
	setIfMissing("postId", "post_id", "id")
	setIfMissing("dictCode", "dict_code", "id")
	setIfMissing("dictId", "dict_id", "id")
	setIfMissing("tableId", "table_id", "id")
	setIfMissing("columnId", "column_id", "id")
	switch strings.TrimSpace(resource) {
	case "sys-users", "sys-user":
		setIfMissing("nickName", "nick_name")
		setIfMissing("roleId", "role_id")
		setIfMissing("deptId", "dept_id")
		setIfMissing("postId", "post_id")
	case "depts":
		setIfMissing("parentId", "parent_id")
		setIfMissing("deptPath", "dept_path")
		setIfMissing("deptName", "dept_name")
	case "posts":
		setIfMissing("postName", "post_name")
		setIfMissing("postCode", "post_code")
	case "apis":
		setIfMissing("id", "apiId", "api_id")
	case "dict-types":
		setIfMissing("dictName", "dict_name")
		setIfMissing("dictType", "dict_type")
	case "dict-data":
		setIfMissing("dictSort", "dict_sort")
		setIfMissing("dictLabel", "dict_label")
		setIfMissing("dictValue", "dict_value")
		setIfMissing("dictType", "dict_type")
		setIfMissing("cssClass", "css_class")
		setIfMissing("listClass", "list_class")
		setIfMissing("isDefault", "is_default")
	case "configs":
		setIfMissing("configName", "config_name")
		setIfMissing("configKey", "config_key")
		setIfMissing("configValue", "config_value")
		setIfMissing("configType", "config_type")
		setIfMissing("isFrontend", "is_frontend")
	}
	return body
}

func (s *serviceImpl) legacySysUserData(ctx context.Context, body map[string]any, creating bool) (any, error) {
	actorID := s.actorID(ctx)
	data := do.SysUser{UpdateBy: actorID}
	if creating {
		data.CreateBy = actorID
	}
	if id := int64Field(body, "userId"); id > 0 {
		data.UserId = id
	}
	setStringDO(body, "username", &data.Username)
	setStringDO(body, "password", &data.Password)
	setStringDO(body, "nickName", &data.NickName)
	setStringDO(body, "phone", &data.Phone)
	setInt64DO(body, "roleId", &data.RoleId)
	setStringDO(body, "avatar", &data.Avatar)
	setStringDO(body, "sex", &data.Sex)
	setStringDO(body, "email", &data.Email)
	setInt64DO(body, "deptId", &data.DeptId)
	setInt64DO(body, "postId", &data.PostId)
	setStringDO(body, "remark", &data.Remark)
	setStringDO(body, "status", &data.Status)
	return data, nil
}

func (s *serviceImpl) legacySysLoginLogData(ctx context.Context, body map[string]any, creating bool) (any, error) {
	actorID := s.actorID(ctx)
	data := do.SysLoginLog{UpdateBy: actorID}
	if creating {
		data.CreateBy = actorID
	}
	setStringDO(body, "username", &data.Username)
	setStringDO(body, "status", &data.Status)
	setStringDO(body, "ipaddr", &data.Ipaddr)
	setStringDO(body, "loginLocation", &data.LoginLocation)
	setStringDO(body, "browser", &data.Browser)
	setStringDO(body, "os", &data.Os)
	setStringDO(body, "platform", &data.Platform)
	setTimeDO(body, "loginTime", &data.LoginTime)
	setStringDO(body, "remark", &data.Remark)
	setStringDO(body, "msg", &data.Msg)
	return data, nil
}

func (s *serviceImpl) legacySysOperaLogData(ctx context.Context, body map[string]any, creating bool) (any, error) {
	actorID := s.actorID(ctx)
	data := do.SysOperaLog{UpdateBy: actorID}
	if creating {
		data.CreateBy = actorID
	}
	setStringDO(body, "title", &data.Title)
	setStringDO(body, "businessType", &data.BusinessType)
	setStringDO(body, "businessTypes", &data.BusinessTypes)
	setStringDO(body, "method", &data.Method)
	setStringDO(body, "requestMethod", &data.RequestMethod)
	setStringDO(body, "operatorType", &data.OperatorType)
	setStringDO(body, "operName", &data.OperName)
	setStringDO(body, "deptName", &data.DeptName)
	setStringDO(body, "operUrl", &data.OperUrl)
	setStringDO(body, "operIp", &data.OperIp)
	setStringDO(body, "operLocation", &data.OperLocation)
	setStringDO(body, "operParam", &data.OperParam)
	setStringDO(body, "status", &data.Status)
	setTimeDO(body, "operTime", &data.OperTime)
	setStringDO(body, "jsonResult", &data.JsonResult)
	setStringDO(body, "remark", &data.Remark)
	setStringDO(body, "latencyTime", &data.LatencyTime)
	setStringDO(body, "userAgent", &data.UserAgent)
	return data, nil
}

func (s *serviceImpl) legacySysAPIData(ctx context.Context, body map[string]any, creating bool) (any, error) {
	actorID := s.actorID(ctx)
	data := do.SysApi{UpdateBy: actorID}
	if creating {
		data.CreateBy = actorID
	}
	if id := int64Field(body, "id"); id > 0 {
		data.Id = id
	}
	setStringDO(body, "handle", &data.Handle)
	setStringDO(body, "title", &data.Title)
	setStringDO(body, "path", &data.Path)
	setStringDO(body, "type", &data.Type)
	setStringDO(body, "action", &data.Action)
	return data, nil
}

func (s *serviceImpl) legacySysDeptData(ctx context.Context, body map[string]any, creating bool) (any, error) {
	actorID := s.actorID(ctx)
	data := do.SysDept{UpdateBy: actorID}
	if creating {
		data.CreateBy = actorID
	}
	if id := int64Field(body, "deptId"); id > 0 {
		data.DeptId = id
	}
	setInt64DO(body, "parentId", &data.ParentId)
	setStringDO(body, "deptPath", &data.DeptPath)
	setStringDO(body, "deptName", &data.DeptName)
	setInt64DO(body, "sort", &data.Sort)
	setStringDO(body, "leader", &data.Leader)
	setStringDO(body, "phone", &data.Phone)
	setStringDO(body, "email", &data.Email)
	setIntDO(body, "status", &data.Status)
	return data, nil
}

func (s *serviceImpl) legacySysPostData(ctx context.Context, body map[string]any, creating bool) (any, error) {
	actorID := s.actorID(ctx)
	data := do.SysPost{UpdateBy: actorID}
	if creating {
		data.CreateBy = actorID
	}
	if id := int64Field(body, "postId"); id > 0 {
		data.PostId = id
	}
	setStringDO(body, "postName", &data.PostName)
	setStringDO(body, "postCode", &data.PostCode)
	setInt64DO(body, "sort", &data.Sort)
	setIntDO(body, "status", &data.Status)
	setStringDO(body, "remark", &data.Remark)
	return data, nil
}

func (s *serviceImpl) legacySysDictTypeData(ctx context.Context, body map[string]any, creating bool) (any, error) {
	actorID := s.actorID(ctx)
	data := do.SysDictType{UpdateBy: actorID}
	if creating {
		data.CreateBy = actorID
	}
	if id := int64Field(body, "dictId"); id > 0 {
		data.DictId = id
	}
	setStringDO(body, "dictName", &data.DictName)
	setStringDO(body, "dictType", &data.DictType)
	setIntDO(body, "status", &data.Status)
	setStringDO(body, "remark", &data.Remark)
	return data, nil
}

func (s *serviceImpl) legacySysDictDataData(ctx context.Context, body map[string]any, creating bool) (any, error) {
	actorID := s.actorID(ctx)
	data := do.SysDictData{UpdateBy: actorID}
	if creating {
		data.CreateBy = actorID
	}
	if id := int64Field(body, "dictCode"); id > 0 {
		data.DictCode = id
	}
	setInt64DO(body, "dictSort", &data.DictSort)
	setStringDO(body, "dictLabel", &data.DictLabel)
	setStringDO(body, "dictValue", &data.DictValue)
	setStringDO(body, "dictType", &data.DictType)
	setStringDO(body, "cssClass", &data.CssClass)
	setStringDO(body, "listClass", &data.ListClass)
	setStringDO(body, "isDefault", &data.IsDefault)
	setIntDO(body, "status", &data.Status)
	setStringDO(body, "default", &data.Default)
	setStringDO(body, "remark", &data.Remark)
	return data, nil
}

func (s *serviceImpl) legacySysConfigData(ctx context.Context, body map[string]any, creating bool) (any, error) {
	actorID := s.actorID(ctx)
	data := do.SysConfig{UpdateBy: actorID}
	if creating {
		data.CreateBy = actorID
	}
	if id := int64Field(body, "id"); id > 0 {
		data.Id = id
	}
	setStringDO(body, "configName", &data.ConfigName)
	setStringDO(body, "configKey", &data.ConfigKey)
	setStringDO(body, "configValue", &data.ConfigValue)
	setStringDO(body, "configType", &data.ConfigType)
	setStringDO(body, "isFrontend", &data.IsFrontend)
	setStringDO(body, "remark", &data.Remark)
	return data, nil
}

func (s *serviceImpl) legacySysRoleData(ctx context.Context, body map[string]any, creating bool) (any, error) {
	actorID := s.actorID(ctx)
	data := do.SysRole{UpdateBy: actorID}
	if creating {
		data.CreateBy = actorID
	}
	if id := int64Field(body, "roleId"); id > 0 {
		data.RoleId = id
	}
	setStringDO(body, "roleName", &data.RoleName)
	setStringDO(body, "status", &data.Status)
	setStringDO(body, "roleKey", &data.RoleKey)
	setInt64DO(body, "roleSort", &data.RoleSort)
	setStringDO(body, "flag", &data.Flag)
	setStringDO(body, "remark", &data.Remark)
	setBoolDO(body, "admin", &data.Admin)
	setStringDO(body, "dataScope", &data.DataScope)
	return data, nil
}

func (s *serviceImpl) legacySysMenuData(ctx context.Context, body map[string]any, creating bool) (any, error) {
	actorID := s.actorID(ctx)
	data := do.SysMenu{UpdateBy: actorID}
	if creating {
		data.CreateBy = actorID
	}
	if id := int64Field(body, "menuId"); id > 0 {
		data.MenuId = id
	}
	setStringDO(body, "menuName", &data.MenuName)
	setStringDO(body, "title", &data.Title)
	setStringDO(body, "icon", &data.Icon)
	setStringDO(body, "path", &data.Path)
	setStringDO(body, "paths", &data.Paths)
	setStringDO(body, "menuType", &data.MenuType)
	setStringDO(body, "action", &data.Action)
	setStringDO(body, "permission", &data.Permission)
	setInt64DO(body, "parentId", &data.ParentId)
	setBoolDO(body, "noCache", &data.NoCache)
	setStringDO(body, "breadcrumb", &data.Breadcrumb)
	setStringDO(body, "component", &data.Component)
	setInt64DO(body, "sort", &data.Sort)
	setStringDO(body, "visible", &data.Visible)
	setStringDO(body, "isFrame", &data.IsFrame)
	return data, nil
}

func (s *serviceImpl) legacySysTablesData(ctx context.Context, body map[string]any, creating bool) (any, error) {
	actorID := s.actorID(ctx)
	data := do.SysTables{UpdateBy: actorID}
	if creating {
		data.CreateBy = actorID
	}
	if id := int64Field(body, "tableId"); id > 0 {
		data.TableId = id
	}
	setStringDO(body, "tableName", &data.TableName)
	setStringDO(body, "tableComment", &data.TableComment)
	setStringDO(body, "className", &data.ClassName)
	setStringDO(body, "tplCategory", &data.TplCategory)
	setStringDO(body, "packageName", &data.PackageName)
	setStringDO(body, "moduleName", &data.ModuleName)
	setStringDO(body, "moduleFrontName", &data.ModuleFrontName)
	setStringDO(body, "businessName", &data.BusinessName)
	setStringDO(body, "functionName", &data.FunctionName)
	setStringDO(body, "functionAuthor", &data.FunctionAuthor)
	setStringDO(body, "remark", &data.Remark)
	return data, nil
}

func (s *serviceImpl) legacySysColumnsData(ctx context.Context, body map[string]any, creating bool) (any, error) {
	actorID := s.actorID(ctx)
	data := do.SysColumns{UpdateBy: actorID}
	if creating {
		data.CreateBy = actorID
	}
	if id := int64Field(body, "columnId"); id > 0 {
		data.ColumnId = id
	}
	setInt64DO(body, "tableId", &data.TableId)
	setStringDO(body, "columnName", &data.ColumnName)
	setStringDO(body, "columnComment", &data.ColumnComment)
	setStringDO(body, "columnType", &data.ColumnType)
	setStringDO(body, "goType", &data.GoType)
	setStringDO(body, "goField", &data.GoField)
	setStringDO(body, "jsonField", &data.JsonField)
	setStringDO(body, "isPk", &data.IsPk)
	setStringDO(body, "isIncrement", &data.IsIncrement)
	setStringDO(body, "isRequired", &data.IsRequired)
	setStringDO(body, "isInsert", &data.IsInsert)
	setStringDO(body, "isEdit", &data.IsEdit)
	setStringDO(body, "isList", &data.IsList)
	setStringDO(body, "isQuery", &data.IsQuery)
	setStringDO(body, "queryType", &data.QueryType)
	setStringDO(body, "htmlType", &data.HtmlType)
	setStringDO(body, "dictType", &data.DictType)
	setInt64DO(body, "sort", &data.Sort)
	setStringDO(body, "remark", &data.Remark)
	return data, nil
}

func setStringDO(body map[string]any, field string, target *any) {
	if hasField(body, field) {
		*target = strings.TrimSpace(gconv.String(body[field]))
	}
}

func setIntDO(body map[string]any, field string, target *any) {
	if hasField(body, field) {
		*target = gconv.Int(body[field])
	}
}

func setInt64DO(body map[string]any, field string, target *any) {
	if hasField(body, field) {
		*target = gconv.Int64(body[field])
	}
}

func setBoolDO(body map[string]any, field string, target *any) {
	if hasField(body, field) {
		*target = gconv.Bool(body[field])
	}
}

func setTimeDO(body map[string]any, field string, target **time.Time) {
	if value := timeField(body, field); value != nil {
		*target = value
	}
}

func buildLegacyRecordTree(records []Record, spec legacyTreeSpec) []Record {
	childrenByParent := make(map[int64][]Record, len(records))
	ids := make(map[int64]struct{}, len(records))
	for _, record := range records {
		id := gconv.Int64(record[spec.idField])
		parentID := gconv.Int64(record[spec.parentField])
		ids[id] = struct{}{}
		childrenByParent[parentID] = append(childrenByParent[parentID], record)
	}
	var attach func(record Record) Record
	attach = func(record Record) Record {
		id := gconv.Int64(record[spec.idField])
		children := childrenByParent[id]
		next := cloneRecord(record)
		if len(children) > 0 {
			nested := make([]Record, 0, len(children))
			for _, child := range children {
				nested = append(nested, attach(child))
			}
			next[spec.childField] = nested
		} else {
			next[spec.childField] = []Record{}
		}
		return next
	}
	result := make([]Record, 0)
	for _, record := range records {
		parentID := gconv.Int64(record[spec.parentField])
		if parentID == 0 {
			result = append(result, attach(record))
			continue
		}
		if _, ok := ids[parentID]; !ok {
			result = append(result, attach(record))
		}
	}
	return result
}

func legacyLabelTreeFromRecords(records []Record, idField string, labelField string) []Record {
	result := make([]Record, 0, len(records))
	for _, record := range records {
		item := Record{"id": record[idField], "label": record[labelField]}
		if children, ok := record["children"].([]Record); ok {
			item["children"] = legacyLabelTreeFromRecords(children, idField, labelField)
		} else {
			item["children"] = []Record{}
		}
		result = append(result, item)
	}
	return result
}

func (s *serviceImpl) legacyMenuLabelTree(ctx context.Context) ([]Record, error) {
	records, err := s.LegacyMenuRoleTree(ctx, 0)
	if err != nil {
		return nil, err
	}
	return legacyLabelTreeFromRecords(records, "menuId", "title"), nil
}

func (s *serviceImpl) legacyRoleMenuIDs(ctx context.Context, roleID int64) ([]int64, error) {
	if roleID <= 0 {
		return []int64{}, nil
	}
	cols := dao.SysRoleMenu.Columns()
	rows, err := dao.SysRoleMenu.Ctx(ctx).Fields(cols.MenuId).Where(cols.RoleId, roleID).All()
	if err != nil {
		return nil, err
	}
	result := make([]int64, 0, len(rows))
	for _, row := range rows {
		if id := row[cols.MenuId].Int64(); id > 0 {
			result = append(result, id)
		}
	}
	return result, nil
}

func (s *serviceImpl) legacyRoleDeptIDs(ctx context.Context, roleID int64) ([]int64, error) {
	if roleID <= 0 {
		return []int64{}, nil
	}
	cols := dao.SysRoleDept.Columns()
	rows, err := dao.SysRoleDept.Ctx(ctx).Fields(cols.DeptId).Where(cols.RoleId, roleID).All()
	if err != nil {
		return nil, err
	}
	result := make([]int64, 0, len(rows))
	for _, row := range rows {
		if id := row[cols.DeptId].Int64(); id > 0 {
			result = append(result, id)
		}
	}
	return result, nil
}

func (s *serviceImpl) legacyRoleRecord(ctx context.Context, roleID int64) (Record, error) {
	if roleID <= 0 {
		return Record{}, nil
	}
	return s.GetLegacySystemResource(ctx, "roles", roleID)
}

func (s *serviceImpl) legacyRolePermissions(ctx context.Context, roleID int64) ([]string, error) {
	menuIDs, err := s.legacyRoleMenuIDs(ctx, roleID)
	if err != nil {
		return nil, err
	}
	if len(menuIDs) == 0 {
		return []string{}, nil
	}
	cols := dao.SysMenu.Columns()
	rows, err := dao.SysMenu.Ctx(ctx).Fields(cols.Permission).WhereIn(cols.MenuId, menuIDs).All()
	if err != nil {
		return nil, err
	}
	result := make([]string, 0, len(rows))
	seen := make(map[string]struct{}, len(rows))
	for _, row := range rows {
		permission := strings.TrimSpace(row[cols.Permission].String())
		if permission == "" {
			continue
		}
		if _, ok := seen[permission]; ok {
			continue
		}
		seen[permission] = struct{}{}
		result = append(result, permission)
	}
	return result, nil
}

func isLegacyAdminRole(roleName string, roleKey string) bool {
	return strings.EqualFold(strings.TrimSpace(roleKey), legacySystemAdminRoleKey) || strings.TrimSpace(roleName) == legacySystemAdminRoleName || strings.EqualFold(strings.TrimSpace(roleName), legacySystemAdminRoleKey)
}

func legacyConfigRowsToMap(rows gdb.Result, keyColumn string, valueColumn string) Record {
	result := make(Record, len(rows))
	for _, row := range rows {
		key := strings.TrimSpace(row[keyColumn].String())
		if key == "" {
			continue
		}
		result[key] = row[valueColumn].String()
	}
	return result
}

func cloneRecord(record Record) Record {
	result := make(Record, len(record))
	for key, value := range record {
		result[key] = value
	}
	return result
}
