// This file wires old go-admin system management HTTP routes to plugin-owned
// compatibility tables. It preserves old response envelopes and request field
// names while delegating all data access to the UIdentity service.

package uidentity

import (
	"fmt"
	"strings"

	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/gogf/gf/v2/util/gconv"
	"golang.org/x/crypto/bcrypt"

	uidentitysvc "lina-plugin-linapro-uidentity-cas/backend/internal/service/uidentity"
)

// LegacySystemList handles old paged system-resource list routes.
func (c *LegacyController) LegacySystemList(resource string) ghttp.HandlerFunc {
	return func(r *ghttp.Request) {
		pageIndex, pageSize := legacyPage(r)
		out, err := c.uidentitySvc.ListLegacySystemResource(r.Context(), uidentitysvc.LegacySystemResourceListInput{
			Resource: resource,
			PageNum:  pageIndex,
			PageSize: pageSize,
			Filters:  legacyRequestMap(r),
			OrderBy:  legacySystemOrderByParam(r),
			Order:    legacySystemOrderParam(r),
		})
		if err != nil {
			legacyError(r, err)
			return
		}
		legacyPageOK(r, out.List, out.Total, pageIndex, pageSize)
	}
}

// LegacySystemGet handles old system-resource detail routes.
func (c *LegacyController) LegacySystemGet(resource string) ghttp.HandlerFunc {
	return func(r *ghttp.Request) {
		record, err := c.uidentitySvc.GetLegacySystemResource(r.Context(), resource, legacySystemRouterID(r))
		if err != nil {
			legacyError(r, err)
			return
		}
		legacyOKWithMsg(r, record, legacyMsgQuerySuccess)
	}
}

// LegacySystemCreate handles old system-resource create routes.
func (c *LegacyController) LegacySystemCreate(resource string) ghttp.HandlerFunc {
	return func(r *ghttp.Request) {
		id, err := c.uidentitySvc.CreateLegacySystemResource(r.Context(), resource, legacyRequestMap(r))
		if err != nil {
			legacyError(r, err)
			return
		}
		legacyOKWithMsg(r, id, legacyMsgCreateSuccess)
	}
}

// LegacySystemUpdate handles old system-resource update routes.
func (c *LegacyController) LegacySystemUpdate(resource string) ghttp.HandlerFunc {
	return func(r *ghttp.Request) {
		id := legacySystemRouterID(r)
		if id <= 0 {
			id = legacyInt64Param(r, "id", "Id", "ID")
		}
		if err := c.uidentitySvc.UpdateLegacySystemResource(r.Context(), resource, id, legacyRequestMap(r)); err != nil {
			legacyError(r, err)
			return
		}
		legacyOKWithMsg(r, id, legacyMsgUpdateSuccess)
	}
}

// LegacySystemDelete handles old system-resource delete routes.
func (c *LegacyController) LegacySystemDelete(resource string) ghttp.HandlerFunc {
	return func(r *ghttp.Request) {
		if err := c.uidentitySvc.DeleteLegacySystemResource(r.Context(), resource, legacySystemDeleteIDs(r)); err != nil {
			legacyError(r, err)
			return
		}
		legacyOKWithMsg(r, legacySystemDeleteIDPayload(r), legacyMsgDeleteSuccess)
	}
}

// LegacyDeptList handles GET /api/v1/dept as an old full-record tree.
func (c *LegacyController) LegacyDeptList(r *ghttp.Request) {
	out, err := c.uidentitySvc.LegacyDeptTree(r.Context(), legacyRequestMap(r))
	if err != nil {
		legacyError(r, err)
		return
	}
	legacyOKWithMsg(r, out, legacyMsgQuerySuccess)
}

// LegacyDeptTree handles GET /api/v1/deptTree as a label tree.
func (c *LegacyController) LegacyDeptTree(r *ghttp.Request) {
	out, err := c.uidentitySvc.LegacyDeptLabelTree(r.Context(), legacyRequestMap(r))
	if err != nil {
		legacyError(r, err)
		return
	}
	legacyOKWithMsg(r, out, legacyMsgQuerySuccess)
}

// LegacyMenuList handles GET /api/v1/menu as the old full-record menu tree.
func (c *LegacyController) LegacyMenuList(r *ghttp.Request) {
	out, err := c.uidentitySvc.LegacyMenuTree(r.Context(), legacyRequestMap(r))
	if err != nil {
		legacyError(r, err)
		return
	}
	legacyOKWithMsg(r, out, legacyMsgQuerySuccess)
}

// LegacyRoleDeptTreeselect handles GET /api/v1/roleDeptTreeselect/{roleId}.
func (c *LegacyController) LegacyRoleDeptTreeselect(r *ghttp.Request) {
	out, err := c.uidentitySvc.LegacyRoleDeptTreeSelect(r.Context(), legacyInt64Param(r, "roleId", "role_id"))
	if err != nil {
		legacyError(r, err)
		return
	}
	legacyOKWithMsg(r, out, legacyMsgQuerySuccess)
}

// LegacyMenuRole handles GET /api/v1/menurole.
func (c *LegacyController) LegacyMenuRole(r *ghttp.Request) {
	out, err := c.uidentitySvc.LegacyMenuRoleTree(r.Context(), legacyInt64Param(r, "roleId", "role_id"))
	if err != nil {
		legacyError(r, err)
		return
	}
	legacyOKWithMsg(r, out, legacyMsgQuerySuccess)
}

// LegacyRoleMenuTreeselect handles GET /api/v1/roleMenuTreeselect/{roleId}.
func (c *LegacyController) LegacyRoleMenuTreeselect(r *ghttp.Request) {
	out, err := c.uidentitySvc.LegacyRoleMenuTreeSelect(r.Context(), legacyInt64Param(r, "roleId", "role_id"))
	if err != nil {
		legacyError(r, err)
		return
	}
	legacyOKWithMsg(r, out, legacyMsgQuerySuccess)
}

// LegacyRoleStatus handles PUT /api/v1/role-status.
func (c *LegacyController) LegacyRoleStatus(r *ghttp.Request) {
	roleID := legacyInt64Param(r, "roleId", "role_id", "id")
	if err := c.uidentitySvc.UpdateLegacySysRoleStatus(r.Context(), roleID, legacyStringParam(r, "status")); err != nil {
		legacyErrorWithMsg(r, err, fmt.Sprintf("更新角色状态失败，失败原因：%s ", err.Error()))
		return
	}
	legacyOKWithMsg(r, roleID, fmt.Sprintf("更新角色 %v 状态成功！", roleID))
}

// LegacyRoleDataScope handles PUT /api/v1/roledatascope.
func (c *LegacyController) LegacyRoleDataScope(r *ghttp.Request) {
	roleID := legacyInt64Param(r, "roleId", "role_id", "id")
	deptIDs := legacyInt64ListParam(r, "deptIds", "dept_ids")
	if err := c.uidentitySvc.UpdateLegacySysRoleDataScope(r.Context(), roleID, legacyStringParam(r, "dataScope", "data_scope"), deptIDs); err != nil {
		legacyErrorWithMsg(r, err, fmt.Sprintf("更新角色数据权限失败！错误详情：%s", err.Error()))
		return
	}
	legacyOKWithMsg(r, nil, legacyMsgOK)
}

// LegacyDictTypeOptions handles GET /api/v1/dict/type-option-select.
func (c *LegacyController) LegacyDictTypeOptions(r *ghttp.Request) {
	out, err := c.uidentitySvc.LegacyDictTypeOptions(r.Context())
	if err != nil {
		legacyError(r, err)
		return
	}
	legacyOKWithMsg(r, out, legacyMsgQuerySuccess)
}

// LegacyDictDataOptions handles GET /api/v1/dict-data/option-select.
func (c *LegacyController) LegacyDictDataOptions(r *ghttp.Request) {
	out, err := c.uidentitySvc.LegacyDictDataOptions(r.Context(), legacyStringParam(r, "dictType", "dict_type"))
	if err != nil {
		legacyError(r, err)
		return
	}
	legacyOKWithMsg(r, out, legacyMsgQuerySuccess)
}

// LegacyConfigByKey handles GET /api/v1/configKey/{configKey}.
func (c *LegacyController) LegacyConfigByKey(r *ghttp.Request) {
	out, err := c.uidentitySvc.LegacyConfigByKey(r.Context(), legacyStringParam(r, "configKey", "config_key"))
	if err != nil {
		legacyError(r, err)
		return
	}
	legacyOKWithMsg(r, out, legacyMsgQuerySuccess)
}

// LegacyAppConfig handles GET /api/v1/app-config.
func (c *LegacyController) LegacyAppConfig(r *ghttp.Request) {
	out, err := c.uidentitySvc.LegacyFrontendConfigs(r.Context())
	if err != nil {
		legacyError(r, err)
		return
	}
	legacyOKWithMsg(r, out, legacyMsgQuerySuccess)
}

// LegacySetConfig handles GET /api/v1/set-config.
func (c *LegacyController) LegacySetConfig(r *ghttp.Request) {
	out, err := c.uidentitySvc.LegacySetConfigs(r.Context())
	if err != nil {
		legacyError(r, err)
		return
	}
	legacyOKWithMsg(r, out, legacyMsgQuerySuccess)
}

// LegacySetConfigUpdate handles PUT /api/v1/set-config.
func (c *LegacyController) LegacySetConfigUpdate(r *ghttp.Request) {
	values := make(map[string]string, len(legacyRequestMap(r)))
	for key, value := range legacyRequestMap(r) {
		if strings.TrimSpace(key) == "" {
			continue
		}
		values[key] = gconv.String(value)
	}
	if err := c.uidentitySvc.UpdateLegacySetConfigs(r.Context(), values); err != nil {
		legacyError(r, err)
		return
	}
	legacyOKWithMsg(r, "", "更新成功")
}

// LegacySysTablesTree handles GET /api/v1/gen/tabletree.
func (c *LegacyController) LegacySysTablesTree(r *ghttp.Request) {
	out, err := c.uidentitySvc.LegacySysTablesTree(r.Context(), legacyRequestMap(r))
	if err != nil {
		legacyError(r, err)
		return
	}
	legacyOKWithMsg(r, out, "")
}

// LegacyGenPreview handles old GET /api/v1/gen/preview/{tableId}.
func (c *LegacyController) LegacyGenPreview(r *ghttp.Request) {
	out, err := c.uidentitySvc.LegacyGenPreview(r.Context(), legacyRouterID(r))
	if err != nil {
		legacyError(r, err)
		return
	}
	legacyOKWithMsg(r, out, "")
}

// LegacyGenToProject handles old GET /api/v1/gen/toproject/{tableId}.
func (c *LegacyController) LegacyGenToProject(r *ghttp.Request) {
	_, err := c.uidentitySvc.LegacyGenToProject(r.Context(), legacyRouterID(r))
	if err != nil {
		legacyError(r, err)
		return
	}
	legacyOKWithMsg(r, "", "Code generated successfully！")
}

// LegacyGenAPIToFile handles old GET /api/v1/gen/apitofile/{tableId}.
func (c *LegacyController) LegacyGenAPIToFile(r *ghttp.Request) {
	_, err := c.uidentitySvc.LegacyGenAPIToFile(r.Context(), legacyRouterID(r))
	if err != nil {
		legacyError(r, err)
		return
	}
	legacyOKWithMsg(r, "", "Code generated successfully！")
}

// LegacyGenToDB handles old GET /api/v1/gen/todb/{tableId}.
func (c *LegacyController) LegacyGenToDB(r *ghttp.Request) {
	_, err := c.uidentitySvc.LegacyGenToDB(r.Context(), legacyRouterID(r))
	if err != nil {
		legacyError(r, err)
		return
	}
	legacyOKWithMsg(r, "", "数据生成成功！")
}

// LegacyGetInfo handles GET /api/v1/getinfo.
func (c *LegacyController) LegacyGetInfo(r *ghttp.Request) {
	out, err := c.uidentitySvc.LegacyGetInfo(r.Context())
	if err != nil {
		legacyError(r, err)
		return
	}
	legacyOKWithMsg(r, out, "")
}

// LegacySystemProfile handles GET /api/v1/user/profile.
func (c *LegacyController) LegacySystemProfile(r *ghttp.Request) {
	out, err := c.uidentitySvc.LegacySystemProfile(r.Context())
	if err != nil {
		legacyErrorWithMsg(r, err, "获取用户信息失败")
		return
	}
	legacyOKWithMsg(r, out, legacyMsgQuerySuccess)
}

// LegacyUserAvatar handles POST /api/v1/user/avatar.
func (c *LegacyController) LegacyUserAvatar(r *ghttp.Request) {
	out, err := c.uidentitySvc.UploadLegacyFiles(r.Context(), uidentitysvc.LegacyUploadInput{
		Type:        "single",
		UploadFiles: r.GetUploadFiles("upload[]"),
	})
	if err != nil {
		legacyError(r, err)
		return
	}
	path := ""
	if out != nil && len(out.Files) > 0 {
		path = out.Files[0].Path
	}
	if err := c.uidentitySvc.UpdateLegacySysUserAvatar(r.Context(), c.legacyCurrentUserID(r), "/"+strings.TrimLeft(path, "/")); err != nil {
		legacyError(r, err)
		return
	}
	legacyOKWithMsg(r, path, "修改成功")
}

// LegacyUserPasswordSet handles PUT /api/v1/user/pwd/set.
func (c *LegacyController) LegacyUserPasswordSet(r *ghttp.Request) {
	hash, err := bcrypt.GenerateFromPassword([]byte(legacyStringParam(r, "newPassword", "new_password", "password")), bcrypt.DefaultCost)
	if err != nil {
		legacyError(r, err)
		return
	}
	if err := c.uidentitySvc.UpdateLegacySysUserPassword(r.Context(), c.legacyCurrentUserID(r), legacyStringParam(r, "oldPassword", "old_password"), string(hash), true); err != nil {
		legacyErrorWithMsg(r, err, "密码修改失败")
		return
	}
	legacyOKWithMsg(r, nil, "密码修改成功")
}

// LegacyUserPasswordReset handles PUT /api/v1/user/pwd/reset.
func (c *LegacyController) LegacyUserPasswordReset(r *ghttp.Request) {
	hash, err := bcrypt.GenerateFromPassword([]byte(legacyStringParam(r, "password", "newPassword", "new_password")), bcrypt.DefaultCost)
	if err != nil {
		legacyError(r, err)
		return
	}
	userID := legacyInt64Param(r, "userId", "user_id", "id")
	if err := c.uidentitySvc.UpdateLegacySysUserPassword(r.Context(), userID, "", string(hash), false); err != nil {
		legacyError(r, err)
		return
	}
	legacyOKWithMsg(r, userID, "更新成功")
}

// LegacyUserStatus handles PUT /api/v1/user/status.
func (c *LegacyController) LegacyUserStatus(r *ghttp.Request) {
	userID := legacyInt64Param(r, "userId", "user_id", "id")
	if err := c.uidentitySvc.UpdateLegacySysUserStatus(r.Context(), userID, legacyStringParam(r, "status")); err != nil {
		legacyError(r, err)
		return
	}
	legacyOKWithMsg(r, userID, "更新成功")
}

func (c *LegacyController) legacyCurrentUserID(r *ghttp.Request) int64 {
	if id := legacyInt64Param(r, "userId", "user_id", "id"); id > 0 {
		return id
	}
	return legacyInt64Param(r, "userID", "uid")
}

func legacySystemRouterID(r *ghttp.Request) int64 {
	if id := legacyRouterID(r); id > 0 {
		return id
	}
	for _, name := range []string{"id", "tableId", "columnId", "dictCode", "dictId", "userId", "deptId", "postId"} {
		if id := legacyInt64Param(r, name); id > 0 {
			return id
		}
	}
	return 0
}

func legacySystemDeleteIDs(r *ghttp.Request) string {
	if id := legacySystemRouterID(r); id > 0 {
		return gconv.String(id)
	}
	return legacyDeleteIDs(r)
}

func legacySystemDeleteIDPayload(r *ghttp.Request) any {
	if id := legacySystemRouterID(r); id > 0 {
		return id
	}
	return legacyDeleteIDPayload(r)
}

func legacySystemOrderParam(r *ghttp.Request) string {
	for _, name := range append(legacySystemOrderFieldNames(), "order", "sortOrder", "sort_order") {
		if value := legacyStringParam(r, name); value != "" {
			return value
		}
	}
	return ""
}

func legacySystemOrderByParam(r *ghttp.Request) string {
	if value := legacyStringParam(r, "orderBy", "order_by", "sort"); value != "" {
		return value
	}
	for _, name := range legacySystemOrderFieldNames() {
		if legacyStringParam(r, name) != "" {
			return strings.TrimSuffix(name, "Order")
		}
	}
	return ""
}

func legacySystemOrderFieldNames() []string {
	return append(legacyResourceOrderFieldNames(),
		"userIdOrder", "usernameOrder", "configNameOrder", "configKeyOrder", "configTypeOrder",
		"titleOrder", "pathOrder", "postNameOrder", "postCodeOrder", "dictIdOrder", "dictTypeOrder",
		"tableIdOrder", "tableNameOrder", "columnIdOrder", "columnNameOrder",
	)
}
