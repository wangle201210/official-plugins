// This file verifies old admin system compatibility methods that previously
// returned empty placeholder payloads.

package uidentity

import (
	"context"
	"fmt"
	"reflect"
	"testing"
	"time"

	"golang.org/x/crypto/bcrypt"

	_ "lina-core/pkg/dbdriver"
	plugincontract "lina-core/pkg/plugin/capability/contract"
	"lina-plugin-linapro-uidentity-cas/backend/internal/dao"
	"lina-plugin-linapro-uidentity-cas/backend/internal/model/do"
)

func TestLegacySystemConfigAndDictOptionsReadPluginTables(t *testing.T) {
	ctx := context.Background()
	configureUIdentityTestDB(t, ctx, dao.SysConfig.Table(), dao.SysDictType.Table(), dao.SysDictData.Table())
	service := &serviceImpl{tenantFilter: testTenantFilter{}}
	suffix := fmt.Sprintf("%d", time.Now().UnixNano())
	configKey := "legacy.system." + suffix
	dictType := "legacy_status_" + suffix
	cleanupLegacySystemRows(t, ctx, suffix)
	t.Cleanup(func() { cleanupLegacySystemRows(t, ctx, suffix) })

	if _, err := dao.SysConfig.Ctx(ctx).Data(do.SysConfig{
		ConfigName:  "Legacy System " + suffix,
		ConfigKey:   configKey,
		ConfigValue: "enabled",
		ConfigType:  "Y",
		IsFrontend:  "1",
	}).Insert(); err != nil {
		t.Fatalf("insert sys config: %v", err)
	}
	if _, err := dao.SysDictType.Ctx(ctx).Data(do.SysDictType{
		DictName: "Legacy Status " + suffix,
		DictType: dictType,
		Status:   1,
	}).Insert(); err != nil {
		t.Fatalf("insert dict type: %v", err)
	}
	if _, err := dao.SysDictData.Ctx(ctx).Data(do.SysDictData{
		DictSort:  1,
		DictLabel: "Enabled",
		DictValue: "1",
		DictType:  dictType,
		Status:    1,
	}).Insert(); err != nil {
		t.Fatalf("insert dict data: %v", err)
	}

	config, err := service.LegacyConfigByKey(ctx, configKey)
	if err != nil {
		t.Fatalf("LegacyConfigByKey: %v", err)
	}
	if config["configKey"] != configKey || config["configValue"] != "enabled" {
		t.Fatalf("unexpected config payload: %#v", config)
	}

	frontend, err := service.LegacyFrontendConfigs(ctx)
	if err != nil {
		t.Fatalf("LegacyFrontendConfigs: %v", err)
	}
	if frontend[configKey] != "enabled" {
		t.Fatalf("frontend config missing: %#v", frontend)
	}

	options, err := service.LegacyDictDataOptions(ctx, dictType)
	if err != nil {
		t.Fatalf("LegacyDictDataOptions: %v", err)
	}
	if len(options) != 1 || options[0]["label"] != "Enabled" || options[0]["value"] != "1" {
		t.Fatalf("unexpected dict data options: %#v", options)
	}
}

func TestLegacySystemDeptAndMenuTreesUseCompatibilityTables(t *testing.T) {
	ctx := context.Background()
	configureUIdentityTestDB(t, ctx, dao.SysDept.Table(), dao.SysRoleDept.Table(), dao.SysMenu.Table(), dao.SysRoleMenu.Table())
	service := &serviceImpl{tenantFilter: testTenantFilter{}}
	suffix := fmt.Sprintf("%d", time.Now().UnixNano())
	cleanupLegacySystemRows(t, ctx, suffix)
	t.Cleanup(func() { cleanupLegacySystemRows(t, ctx, suffix) })

	rootDept := insertLegacySystemDept(t, ctx, 0, "Legacy Root "+suffix, 1)
	childDept := insertLegacySystemDept(t, ctx, rootDept, "Legacy Child "+suffix, 2)
	if _, err := dao.SysRoleDept.Ctx(ctx).Data(do.SysRoleDept{RoleId: 7711, DeptId: childDept}).Insert(); err != nil {
		t.Fatalf("insert role dept: %v", err)
	}
	rootMenu := insertLegacySystemMenu(t, ctx, 0, "Legacy Menu "+suffix, "legacy:menu:"+suffix, 1)
	childMenu := insertLegacySystemMenu(t, ctx, rootMenu, "Legacy Child Menu "+suffix, "legacy:child:"+suffix, 2)
	if _, err := dao.SysRoleMenu.Ctx(ctx).Data(do.SysRoleMenu{RoleId: 7711, MenuId: childMenu}).Insert(); err != nil {
		t.Fatalf("insert role menu: %v", err)
	}

	deptTree, err := service.LegacyDeptLabelTree(ctx, map[string]any{"deptName": "Legacy"})
	if err != nil {
		t.Fatalf("LegacyDeptLabelTree: %v", err)
	}
	if !legacyTreeContainsLabel(deptTree, "Legacy Child "+suffix) {
		t.Fatalf("dept label tree missing child: %#v", deptTree)
	}
	roleDept, err := service.LegacyRoleDeptTreeSelect(ctx, 7711)
	if err != nil {
		t.Fatalf("LegacyRoleDeptTreeSelect: %v", err)
	}
	if !reflect.DeepEqual(roleDept["checkedKeys"], []int64{childDept}) {
		t.Fatalf("unexpected role dept checked keys: %#v", roleDept)
	}

	roleMenu, err := service.LegacyRoleMenuTreeSelect(ctx, 7711)
	if err != nil {
		t.Fatalf("LegacyRoleMenuTreeSelect: %v", err)
	}
	if !reflect.DeepEqual(roleMenu["checkedKeys"], []int64{childMenu}) {
		t.Fatalf("unexpected role menu checked keys: %#v", roleMenu)
	}
}

func TestLegacySystemUserListFiltersByOldDeptPath(t *testing.T) {
	ctx := context.Background()
	configureUIdentityTestDB(t, ctx, dao.SysUser.Table(), dao.SysDept.Table())
	service := &serviceImpl{tenantFilter: testTenantFilter{}}
	suffix := fmt.Sprintf("%d", time.Now().UnixNano())
	cleanupLegacySystemRows(t, ctx, suffix)
	t.Cleanup(func() { cleanupLegacySystemRows(t, ctx, suffix) })

	rootDept := insertLegacySystemDept(t, ctx, 0, "Legacy User Root "+suffix, 1)
	childDept := insertLegacySystemDept(t, ctx, rootDept, "Legacy User Child "+suffix, 2)
	if _, err := dao.SysDept.Ctx(ctx).
		Where(dao.SysDept.Columns().DeptId, rootDept).
		OmitNilData().
		Data(do.SysDept{DeptPath: fmt.Sprintf("0,%d", rootDept)}).
		Update(); err != nil {
		t.Fatalf("update root dept path: %v", err)
	}
	if _, err := dao.SysDept.Ctx(ctx).
		Where(dao.SysDept.Columns().DeptId, childDept).
		OmitNilData().
		Data(do.SysDept{DeptPath: fmt.Sprintf("0,%d,%d", rootDept, childDept)}).
		Update(); err != nil {
		t.Fatalf("update child dept path: %v", err)
	}
	if _, err := dao.SysUser.Ctx(ctx).Data(do.SysUser{
		Username: "legacy-path-child-" + suffix,
		NickName: "Legacy Path Child " + suffix,
		DeptId:   childDept,
		Status:   "1",
	}).Insert(); err != nil {
		t.Fatalf("insert child user: %v", err)
	}
	if _, err := dao.SysUser.Ctx(ctx).Data(do.SysUser{
		Username: "legacy-path-other-" + suffix,
		NickName: "Legacy Path Other " + suffix,
		DeptId:   rootDept + childDept + 991,
		Status:   "1",
	}).Insert(); err != nil {
		t.Fatalf("insert other user: %v", err)
	}

	out, err := service.ListLegacySystemResource(ctx, LegacySystemResourceListInput{
		Resource: "sys-users",
		PageNum:  1,
		PageSize: 20,
		Filters:  map[string]any{"deptId": rootDept, "keyword": "legacy-path-"},
	})
	if err != nil {
		t.Fatalf("ListLegacySystemResource: %v", err)
	}
	if out.Total != 1 || len(out.List) != 1 || out.List[0]["username"] != "legacy-path-child-"+suffix {
		t.Fatalf("unexpected dept path filtered users: %#v", out)
	}
}

func TestLegacyGetInfoProjectsOldShapeAndPermissions(t *testing.T) {
	ctx := context.Background()
	configureUIdentityTestDB(t, ctx, dao.SysUser.Table(), dao.SysRole.Table(), dao.SysMenu.Table(), dao.SysRoleMenu.Table())
	actorID := int64(88001 + time.Now().UnixNano()%100000)
	service := &serviceImpl{tenantFilter: testTenantFilter{current: plugincontract.TenantFilterContext{UserID: int(actorID)}}}
	suffix := fmt.Sprintf("%d", time.Now().UnixNano())
	cleanupLegacySystemRows(t, ctx, suffix)
	t.Cleanup(func() { cleanupLegacySystemRows(t, ctx, suffix) })

	roleID, err := dao.SysRole.Ctx(ctx).Data(do.SysRole{
		RoleName: "LegacyRole" + suffix,
		RoleKey:  "legacy_role_" + suffix,
		Status:   "1",
	}).InsertAndGetId()
	if err != nil {
		t.Fatalf("insert role: %v", err)
	}
	menuID := insertLegacySystemMenu(t, ctx, 0, "Legacy Permission Menu "+suffix, "legacy:perm:"+suffix, 1)
	if _, err := dao.SysRoleMenu.Ctx(ctx).Data(do.SysRoleMenu{RoleId: roleID, MenuId: menuID}).Insert(); err != nil {
		t.Fatalf("insert role menu: %v", err)
	}
	password, err := bcrypt.GenerateFromPassword([]byte("old-password"), bcrypt.DefaultCost)
	if err != nil {
		t.Fatalf("hash password: %v", err)
	}
	if _, err := dao.SysUser.Ctx(ctx).Data(do.SysUser{
		UserId:   actorID,
		Username: "legacy-user-" + suffix,
		Password: string(password),
		NickName: "Legacy User " + suffix,
		RoleId:   roleID,
		DeptId:   19,
		Status:   "1",
	}).Insert(); err != nil {
		t.Fatalf("insert user: %v", err)
	}

	info, err := service.LegacyGetInfo(ctx)
	if err != nil {
		t.Fatalf("LegacyGetInfo: %v", err)
	}
	if info["userName"] != "Legacy User "+suffix || info["userId"] != actorID || info["deptId"] != int64(19) {
		t.Fatalf("unexpected getinfo user projection: %#v", info)
	}
	if !reflect.DeepEqual(info["permissions"], []string{"legacy:perm:" + suffix}) {
		t.Fatalf("unexpected permissions: %#v", info)
	}
}

func insertLegacySystemDept(t *testing.T, ctx context.Context, parentID int64, name string, sort int) int64 {
	t.Helper()
	id, err := dao.SysDept.Ctx(ctx).Data(do.SysDept{
		ParentId: parentID,
		DeptName: name,
		Sort:     sort,
		Status:   1,
	}).InsertAndGetId()
	if err != nil {
		t.Fatalf("insert dept %s: %v", name, err)
	}
	return id
}

func insertLegacySystemMenu(t *testing.T, ctx context.Context, parentID int64, title string, permission string, sort int) int64 {
	t.Helper()
	id, err := dao.SysMenu.Ctx(ctx).Data(do.SysMenu{
		Title:      title,
		MenuName:   title,
		ParentId:   parentID,
		Permission: permission,
		MenuType:   "C",
		Sort:       sort,
		Visible:    "0",
	}).InsertAndGetId()
	if err != nil {
		t.Fatalf("insert menu %s: %v", title, err)
	}
	return id
}

func legacyTreeContainsLabel(records []Record, label string) bool {
	for _, record := range records {
		if record["label"] == label {
			return true
		}
		if children, ok := record["children"].([]Record); ok && legacyTreeContainsLabel(children, label) {
			return true
		}
	}
	return false
}

func cleanupLegacySystemRows(t *testing.T, ctx context.Context, suffix string) {
	t.Helper()
	if _, err := dao.SysConfig.Ctx(ctx).Unscoped().WhereLike(dao.SysConfig.Columns().ConfigKey, "%"+suffix).Delete(); err != nil {
		t.Fatalf("cleanup sys config: %v", err)
	}
	if _, err := dao.SysDictData.Ctx(ctx).Unscoped().WhereLike(dao.SysDictData.Columns().DictType, "%"+suffix).Delete(); err != nil {
		t.Fatalf("cleanup sys dict data: %v", err)
	}
	if _, err := dao.SysDictType.Ctx(ctx).Unscoped().WhereLike(dao.SysDictType.Columns().DictType, "%"+suffix).Delete(); err != nil {
		t.Fatalf("cleanup sys dict type: %v", err)
	}
	if _, err := dao.SysRoleDept.Ctx(ctx).Where(dao.SysRoleDept.Columns().RoleId, 7711).Delete(); err != nil {
		t.Fatalf("cleanup sys role dept: %v", err)
	}
	if _, err := dao.SysDept.Ctx(ctx).Unscoped().WhereLike(dao.SysDept.Columns().DeptName, "%"+suffix).Delete(); err != nil {
		t.Fatalf("cleanup sys dept: %v", err)
	}
	if _, err := dao.SysRoleMenu.Ctx(ctx).Where(dao.SysRoleMenu.Columns().RoleId, 7711).Delete(); err != nil {
		t.Fatalf("cleanup fixed sys role menu: %v", err)
	}
	if rows, err := dao.SysRole.Ctx(ctx).Unscoped().Fields(dao.SysRole.Columns().RoleId).WhereLike(dao.SysRole.Columns().RoleKey, "%"+suffix).Array(); err == nil {
		ids := make([]int64, 0, len(rows))
		for _, row := range rows {
			if id := row.Int64(); id > 0 {
				ids = append(ids, id)
			}
		}
		if len(ids) > 0 {
			if _, err := dao.SysRoleMenu.Ctx(ctx).WhereIn(dao.SysRoleMenu.Columns().RoleId, ids).Delete(); err != nil {
				t.Fatalf("cleanup sys role menu: %v", err)
			}
		}
	} else {
		t.Fatalf("query sys roles for cleanup: %v", err)
	}
	if _, err := dao.SysMenu.Ctx(ctx).Unscoped().WhereLike(dao.SysMenu.Columns().Title, "%"+suffix).Delete(); err != nil {
		t.Fatalf("cleanup sys menu: %v", err)
	}
	if _, err := dao.SysUser.Ctx(ctx).Unscoped().WhereLike(dao.SysUser.Columns().Username, "%"+suffix).Delete(); err != nil {
		t.Fatalf("cleanup sys user: %v", err)
	}
	if _, err := dao.SysRole.Ctx(ctx).Unscoped().WhereLike(dao.SysRole.Columns().RoleKey, "%"+suffix).Delete(); err != nil {
		t.Fatalf("cleanup sys role: %v", err)
	}
}
