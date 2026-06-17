// This file verifies old admin system compatibility methods that previously
// returned empty placeholder payloads.

package uidentity

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/gogf/gf/v2/util/gconv"
	"golang.org/x/crypto/bcrypt"

	_ "lina-core/pkg/dbdriver"
	"lina-core/pkg/plugin/capability/tenantcap/tenantspi"
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
	service := &serviceImpl{tenantFilter: testTenantFilter{current: tenantspi.TenantFilterContext{UserID: int(actorID)}}}
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

func TestLegacyAdminLogoutWritesSysLoginLog(t *testing.T) {
	ctx := context.Background()
	configureUIdentityTestDB(t, ctx, dao.SysLoginLog.Table())
	actorID := int64(91001 + time.Now().UnixNano()%100000)
	service := &serviceImpl{tenantFilter: testTenantFilter{current: tenantspi.TenantFilterContext{UserID: int(actorID)}}}
	suffix := fmt.Sprintf("%d", time.Now().UnixNano())
	username := "legacy-logout-" + suffix
	t.Cleanup(func() {
		if _, err := dao.SysLoginLog.Ctx(ctx).Unscoped().Where(dao.SysLoginLog.Columns().Username, username).Delete(); err != nil {
			t.Fatalf("cleanup sys login log: %v", err)
		}
	})

	if err := service.RecordLegacyAdminLogout(ctx, LegacyAdminLogoutInput{
		Username:  username,
		IP:        "127.0.0.1",
		UserAgent: "Mozilla/5.0 (Macintosh) Firefox/120.0",
	}); err != nil {
		t.Fatalf("RecordLegacyAdminLogout: %v", err)
	}
	var row struct {
		Username string `json:"username"`
		Status   string `json:"status"`
		Ipaddr   string `json:"ipaddr"`
		Msg      string `json:"msg"`
		Remark   string `json:"remark"`
		CreateBy int64  `json:"createBy"`
		UpdateBy int64  `json:"updateBy"`
	}
	if err := dao.SysLoginLog.Ctx(ctx).
		Fields("username", "status", "ipaddr", "msg", "remark", "create_by", "update_by").
		Where(dao.SysLoginLog.Columns().Username, username).
		OrderDesc(dao.SysLoginLog.Columns().Id).
		Scan(&row); err != nil {
		t.Fatalf("query sys login log: %v", err)
	}
	if row.Username != username || row.Status != legacyAdminLogoutStatus ||
		row.Ipaddr != "127.0.0.1" || row.Msg != legacyAdminLogoutMessage ||
		row.Remark != "Mozilla/5.0 (Macintosh) Firefox/120.0" ||
		row.CreateBy != actorID || row.UpdateBy != actorID {
		t.Fatalf("unexpected legacy logout log row: %#v", row)
	}
}

func TestLegacyAdminLogoutHonorsLoggerDBSwitch(t *testing.T) {
	ctx := context.Background()
	configureUIdentityTestDB(t, ctx, dao.SysLoginLog.Table())
	service := &serviceImpl{
		configSvc:    newLegacyConfigTestService(t, "legacy:\n  logger:\n    enabledDB: false\n"),
		tenantFilter: testTenantFilter{},
	}
	suffix := fmt.Sprintf("%d", time.Now().UnixNano())
	username := "legacy-logout-disabled-" + suffix
	t.Cleanup(func() {
		if _, err := dao.SysLoginLog.Ctx(ctx).Unscoped().Where(dao.SysLoginLog.Columns().Username, username).Delete(); err != nil {
			t.Fatalf("cleanup disabled sys login log: %v", err)
		}
	})

	if err := service.RecordLegacyAdminLogout(ctx, LegacyAdminLogoutInput{Username: username}); err != nil {
		t.Fatalf("RecordLegacyAdminLogout with disabled logger: %v", err)
	}
	count, err := dao.SysLoginLog.Ctx(ctx).Where(dao.SysLoginLog.Columns().Username, username).Count()
	if err != nil {
		t.Fatalf("count disabled sys login log: %v", err)
	}
	if count != 0 {
		t.Fatalf("disabled legacy logger wrote %d rows for %s", count, username)
	}
}

func TestLegacySystemProfileProjectsUserRolesPosts(t *testing.T) {
	ctx := context.Background()
	configureUIdentityTestDB(t, ctx, dao.SysUser.Table(), dao.SysRole.Table(), dao.SysPost.Table(), dao.SysDept.Table())
	actorID := int64(89001 + time.Now().UnixNano()%100000)
	service := &serviceImpl{tenantFilter: testTenantFilter{current: tenantspi.TenantFilterContext{UserID: int(actorID)}}}
	suffix := fmt.Sprintf("%d", time.Now().UnixNano())
	cleanupLegacySystemRows(t, ctx, suffix)
	t.Cleanup(func() { cleanupLegacySystemRows(t, ctx, suffix) })

	deptID := insertLegacySystemDept(t, ctx, 0, "Legacy Profile Dept "+suffix, 1)
	roleID, err := dao.SysRole.Ctx(ctx).Data(do.SysRole{
		RoleName: "Legacy Profile Role " + suffix,
		RoleKey:  "legacy_profile_" + suffix,
		Status:   "1",
	}).InsertAndGetId()
	if err != nil {
		t.Fatalf("insert role: %v", err)
	}
	postID, err := dao.SysPost.Ctx(ctx).Data(do.SysPost{
		PostName: "Legacy Profile Post " + suffix,
		PostCode: "legacy-profile-" + suffix,
		Status:   1,
	}).InsertAndGetId()
	if err != nil {
		t.Fatalf("insert post: %v", err)
	}
	if _, err := dao.SysUser.Ctx(ctx).Data(do.SysUser{
		UserId:   actorID,
		Username: "legacy-profile-user-" + suffix,
		NickName: "Legacy Profile User " + suffix,
		RoleId:   roleID,
		DeptId:   deptID,
		PostId:   postID,
		Status:   "1",
	}).Insert(); err != nil {
		t.Fatalf("insert user: %v", err)
	}

	profile, err := service.LegacySystemProfile(ctx)
	if err != nil {
		t.Fatalf("LegacySystemProfile: %v", err)
	}
	user, ok := profile["user"].(Record)
	if !ok || user["userId"] != actorID || user["nickName"] != "Legacy Profile User "+suffix {
		t.Fatalf("unexpected profile user: %#v", profile["user"])
	}
	dept, ok := user["dept"].(Record)
	if !ok || dept["deptId"] != deptID || dept["deptName"] != "Legacy Profile Dept "+suffix {
		t.Fatalf("unexpected profile dept: %#v", user["dept"])
	}
	roles, ok := profile["roles"].([]Record)
	if !ok || len(roles) != 1 || roles[0]["roleId"] != roleID || roles[0]["roleName"] != "Legacy Profile Role "+suffix {
		t.Fatalf("unexpected profile roles: %#v", profile["roles"])
	}
	posts, ok := profile["posts"].([]Record)
	if !ok || len(posts) != 1 || posts[0]["postId"] != postID || posts[0]["postName"] != "Legacy Profile Post "+suffix {
		t.Fatalf("unexpected profile posts: %#v", profile["posts"])
	}
}

func TestLegacyRoleActionsUpdateCompatibilityTables(t *testing.T) {
	ctx := context.Background()
	configureUIdentityTestDB(t, ctx, dao.SysRole.Table(), dao.SysDept.Table(), dao.SysRoleDept.Table())
	service := &serviceImpl{tenantFilter: testTenantFilter{current: tenantspi.TenantFilterContext{UserID: 7031}}}
	suffix := fmt.Sprintf("%d", time.Now().UnixNano())
	cleanupLegacySystemRows(t, ctx, suffix)
	t.Cleanup(func() { cleanupLegacySystemRows(t, ctx, suffix) })

	roleID, err := dao.SysRole.Ctx(ctx).Data(do.SysRole{
		RoleName:  "Legacy Action Role " + suffix,
		RoleKey:   "legacy_action_role_" + suffix,
		Status:    "1",
		DataScope: "1",
	}).InsertAndGetId()
	if err != nil {
		t.Fatalf("insert role: %v", err)
	}
	deptA := insertLegacySystemDept(t, ctx, 0, "Legacy Action Dept A "+suffix, 1)
	deptB := insertLegacySystemDept(t, ctx, 0, "Legacy Action Dept B "+suffix, 2)
	if _, err := dao.SysRoleDept.Ctx(ctx).Data(do.SysRoleDept{RoleId: roleID, DeptId: deptA}).Insert(); err != nil {
		t.Fatalf("insert role dept: %v", err)
	}

	if err := service.UpdateLegacySysRoleStatus(ctx, roleID, "0"); err != nil {
		t.Fatalf("UpdateLegacySysRoleStatus: %v", err)
	}
	role, err := service.GetLegacySystemResource(ctx, "roles", roleID)
	if err != nil {
		t.Fatalf("GetLegacySystemResource: %v", err)
	}
	if role["status"] != "0" {
		t.Fatalf("role status not updated: %#v", role)
	}

	if err := service.UpdateLegacySysRoleDataScope(ctx, roleID, "2", []int64{deptB, deptB, -1, 0}); err != nil {
		t.Fatalf("UpdateLegacySysRoleDataScope: %v", err)
	}
	role, err = service.GetLegacySystemResource(ctx, "roles", roleID)
	if err != nil {
		t.Fatalf("GetLegacySystemResource after scope update: %v", err)
	}
	if role["dataScope"] != "2" {
		t.Fatalf("role data scope not updated: %#v", role)
	}
	checked, err := service.legacyRoleDeptIDs(ctx, roleID)
	if err != nil {
		t.Fatalf("legacyRoleDeptIDs: %v", err)
	}
	if !reflect.DeepEqual(checked, []int64{deptB}) {
		t.Fatalf("unexpected role dept IDs: %#v", checked)
	}
}

func TestLegacySysTablesTreeReadsTablesAndColumns(t *testing.T) {
	ctx := context.Background()
	configureUIdentityTestDB(t, ctx, dao.SysTables.Table(), dao.SysColumns.Table())
	service := &serviceImpl{tenantFilter: testTenantFilter{}}
	suffix := fmt.Sprintf("%d", time.Now().UnixNano())
	cleanupLegacySystemRows(t, ctx, suffix)
	t.Cleanup(func() { cleanupLegacySystemRows(t, ctx, suffix) })

	tableID, err := dao.SysTables.Ctx(ctx).Data(do.SysTables{
		TableName:    "legacy_table_" + suffix,
		TableComment: "Legacy Table " + suffix,
		ClassName:    "LegacyTable" + suffix,
	}).InsertAndGetId()
	if err != nil {
		t.Fatalf("insert sys table: %v", err)
	}
	if _, err := dao.SysColumns.Ctx(ctx).Data(do.SysColumns{
		TableId:       tableID,
		ColumnName:    "legacy_column_" + suffix,
		ColumnComment: "Legacy Column " + suffix,
		GoField:       "LegacyColumn",
		JsonField:     "legacyColumn",
		Sort:          1,
	}).Insert(); err != nil {
		t.Fatalf("insert sys column: %v", err)
	}

	tree, err := service.LegacySysTablesTree(ctx, map[string]any{"tableName": "legacy_table_" + suffix})
	if err != nil {
		t.Fatalf("LegacySysTablesTree: %v", err)
	}
	if len(tree) != 1 || tree[0]["tableName"] != "legacy_table_"+suffix {
		t.Fatalf("unexpected table tree rows: %#v", tree)
	}
	columns, ok := tree[0]["columns"].([]Record)
	if !ok || len(columns) != 1 {
		t.Fatalf("table columns missing: %#v", tree[0]["columns"])
	}
	if columns[0]["columnName"] != "legacy_column_"+suffix || columns[0]["jsonField"] != "legacyColumn" {
		t.Fatalf("unexpected table column projection: %#v", columns)
	}
}

func TestLegacyGenRoutesRenderFilesAndMenus(t *testing.T) {
	ctx := context.Background()
	configureUIdentityTestDB(t, ctx, dao.SysTables.Table(), dao.SysColumns.Table(), dao.SysMenu.Table())
	suffix := fmt.Sprintf("%d", time.Now().UnixNano())
	cleanupLegacySystemRows(t, ctx, suffix)
	t.Cleanup(func() { cleanupLegacySystemRows(t, ctx, suffix) })

	tableID := insertLegacyGenTable(t, ctx, suffix)
	genRoot := t.TempDir()
	frontRoot := t.TempDir()
	service := &serviceImpl{
		configSvc: newLegacyConfigTestService(t, fmt.Sprintf(`
legacy:
  gen:
    outputRoot: %q
    frontOutputRoot: %q
`, genRoot, frontRoot)),
		tenantFilter: testTenantFilter{},
	}

	preview, err := service.LegacyGenPreview(ctx, tableID)
	if err != nil {
		t.Fatalf("LegacyGenPreview: %v", err)
	}
	if got := gconv.String(preview["template/model.go.template"]); !strings.Contains(got, "type LegacyGen"+suffix+" struct") || !strings.Contains(got, `return "legacy_gen_`+suffix+`"`) {
		t.Fatalf("unexpected model preview: %s", got)
	}
	if got := gconv.String(preview["template/api.go.template"]); !strings.Contains(got, "GetPage") || !strings.Contains(got, "LegacyGen"+suffix) {
		t.Fatalf("unexpected api preview: %s", got)
	}

	generated, err := service.LegacyGenToProject(ctx, tableID)
	if err != nil {
		t.Fatalf("LegacyGenToProject: %v", err)
	}
	paths := generated["paths"].([]string)
	if len(paths) != 7 {
		t.Fatalf("generated paths = %#v", paths)
	}
	assertLegacyGenFileContains(t, filepath.Join(genRoot, "app", "legacy", "models", "legacy_gen_"+suffix+".go"), "LegacyGen"+suffix)
	assertLegacyGenFileContains(t, filepath.Join(frontRoot, "api", "legacy", "legacy-gen-"+suffix+".js"), "/api/v1/legacy_gen_"+suffix)

	apiFile, err := service.LegacyGenAPIToFile(ctx, tableID)
	if err != nil {
		t.Fatalf("LegacyGenAPIToFile: %v", err)
	}
	if path := gconv.String(apiFile["path"]); !strings.HasPrefix(path, filepath.Join(genRoot, "cmd", "migrate", "migration", "version-local")) {
		t.Fatalf("unexpected api migrate path: %s", path)
	} else {
		assertLegacyGenFileContains(t, path, "migration.Migrate.SetVersion")
	}

	menuOut, err := service.LegacyGenToDB(ctx, tableID)
	if err != nil {
		t.Fatalf("LegacyGenToDB: %v", err)
	}
	if gconv.Int(menuOut["count"]) != 12 {
		t.Fatalf("unexpected generated menu count: %#v", menuOut)
	}
	count, err := dao.SysMenu.Ctx(ctx).WhereLike(dao.SysMenu.Columns().Title, "%Legacy Generated "+suffix+"%").Count()
	if err != nil {
		t.Fatalf("count generated menus: %v", err)
	}
	if count != 12 {
		t.Fatalf("generated menu row count = %d, want 12", count)
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

func insertLegacyGenTable(t *testing.T, ctx context.Context, suffix string) int64 {
	t.Helper()
	tableID, err := dao.SysTables.Ctx(ctx).Data(do.SysTables{
		TableName:       "legacy_gen_" + suffix,
		TableComment:    "Legacy Generated " + suffix,
		ClassName:       "LegacyGen" + suffix,
		PackageName:     "legacy",
		ModuleName:      "legacy_gen_" + suffix,
		ModuleFrontName: "legacy-gen-" + suffix,
		BusinessName:    "legacyGen" + suffix,
		FunctionName:    "Legacy Generated " + suffix,
		FunctionAuthor:  "codex",
		PkColumn:        "id",
		PkGoField:       "Id",
		PkJsonField:     "id",
	}).InsertAndGetId()
	if err != nil {
		t.Fatalf("insert gen table: %v", err)
	}
	columns := []do.SysColumns{
		{
			TableId:       tableID,
			ColumnName:    "id",
			ColumnComment: "ID",
			ColumnType:    "bigint",
			GoType:        "int",
			GoField:       "Id",
			JsonField:     "id",
			IsPk:          "1",
			IsIncrement:   "1",
			Pk:            true,
			Increment:     true,
			Sort:          1,
		},
		{
			TableId:       tableID,
			ColumnName:    "name",
			ColumnComment: "名称",
			ColumnType:    "varchar(128)",
			GoType:        "string",
			GoField:       "Name",
			JsonField:     "name",
			IsInsert:      "1",
			IsEdit:        "1",
			IsList:        "1",
			IsQuery:       "1",
			QueryType:     "LIKE",
			HtmlType:      "input",
			Insert:        true,
			Edit:          true,
			Query:         true,
			List:          "1",
			Sort:          2,
		},
	}
	for _, column := range columns {
		if _, err := dao.SysColumns.Ctx(ctx).Data(column).Insert(); err != nil {
			t.Fatalf("insert gen column: %v", err)
		}
	}
	return tableID
}

func assertLegacyGenFileContains(t *testing.T, path string, want string) {
	t.Helper()
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read generated file %s: %v", path, err)
	}
	if !strings.Contains(string(content), want) {
		t.Fatalf("generated file %s does not contain %q: %s", path, want, string(content))
	}
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
	if rows, err := dao.SysRole.Ctx(ctx).Unscoped().Fields(dao.SysRole.Columns().RoleId).WhereLike(dao.SysRole.Columns().RoleKey, "%"+suffix).Array(); err == nil {
		ids := make([]int64, 0, len(rows))
		for _, row := range rows {
			if id := row.Int64(); id > 0 {
				ids = append(ids, id)
			}
		}
		if len(ids) > 0 {
			if _, err := dao.SysRoleDept.Ctx(ctx).WhereIn(dao.SysRoleDept.Columns().RoleId, ids).Delete(); err != nil {
				t.Fatalf("cleanup dynamic sys role dept: %v", err)
			}
		}
	} else {
		t.Fatalf("query sys roles for dept cleanup: %v", err)
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
	if rows, err := dao.SysTables.Ctx(ctx).Unscoped().Fields(dao.SysTables.Columns().TableId).WhereLike(dao.SysTables.Columns().TableName, "%"+suffix).Array(); err == nil {
		ids := make([]int64, 0, len(rows))
		for _, row := range rows {
			if id := row.Int64(); id > 0 {
				ids = append(ids, id)
			}
		}
		if len(ids) > 0 {
			if _, err := dao.SysColumns.Ctx(ctx).Unscoped().WhereIn(dao.SysColumns.Columns().TableId, ids).Delete(); err != nil {
				t.Fatalf("cleanup sys columns: %v", err)
			}
		}
	} else {
		t.Fatalf("query sys tables for cleanup: %v", err)
	}
	if _, err := dao.SysTables.Ctx(ctx).Unscoped().WhereLike(dao.SysTables.Columns().TableName, "%"+suffix).Delete(); err != nil {
		t.Fatalf("cleanup sys tables: %v", err)
	}
	if _, err := dao.SysRole.Ctx(ctx).Unscoped().WhereLike(dao.SysRole.Columns().RoleKey, "%"+suffix).Delete(); err != nil {
		t.Fatalf("cleanup sys role: %v", err)
	}
}
