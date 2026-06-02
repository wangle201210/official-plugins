// This file adapts old uidentity/admin routes that collide with LinaPro host
// system-management routes. It lets the host run authentication, tenancy,
// permission, and real system-management handlers first, then rewrites successful
// host JSON responses back to the old GoAdmin response envelope.

package backend

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/gogf/gf/v2/util/gconv"

	"lina-core/pkg/plugin/pluginhost"
)

const (
	legacyAPIV1Prefix        = "/api/v1"
	legacyHostSuccessCode    = 0
	legacyHTTPStatusOK       = 200
	legacyQuerySuccessMsg    = "查询成功"
	legacyCreateSuccessMsg   = "创建成功"
	legacyUpdateSuccessMsg   = "更新成功"
	legacyDeleteSuccessMsg   = "删除成功"
	legacyAPIV1ScopePattern  = "/api/v1/*"
	legacyBearerHeaderPrefix = "Bearer "
)

type legacyHostRouteKind string

const (
	legacyHostRouteList   legacyHostRouteKind = "list"
	legacyHostRouteTree   legacyHostRouteKind = "tree"
	legacyHostRouteDetail legacyHostRouteKind = "detail"
	legacyHostRouteCreate legacyHostRouteKind = "create"
	legacyHostRouteUpdate legacyHostRouteKind = "update"
	legacyHostRouteDelete legacyHostRouteKind = "delete"
)

type legacyHostRouteAdapter struct {
	Method string
	Path   string
	Kind   legacyHostRouteKind
}

type legacyHostResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Msg     string `json:"msg"`
	Data    any    `json:"data"`
}

var legacyHostRouteAdapters = []legacyHostRouteAdapter{
	{Method: "GET", Path: "/user/profile", Kind: legacyHostRouteDetail},
	{Method: "GET", Path: "/role", Kind: legacyHostRouteList},
	{Method: "GET", Path: "/role/{id}", Kind: legacyHostRouteDetail},
	{Method: "POST", Path: "/role", Kind: legacyHostRouteCreate},
	{Method: "PUT", Path: "/role/{id}", Kind: legacyHostRouteUpdate},
	{Method: "DELETE", Path: "/role", Kind: legacyHostRouteDelete},
	{Method: "GET", Path: "/menu", Kind: legacyHostRouteTree},
	{Method: "GET", Path: "/menu/{id}", Kind: legacyHostRouteDetail},
	{Method: "POST", Path: "/menu", Kind: legacyHostRouteCreate},
	{Method: "PUT", Path: "/menu/{id}", Kind: legacyHostRouteUpdate},
	{Method: "GET", Path: "/dict/data", Kind: legacyHostRouteList},
	{Method: "GET", Path: "/dict/data/{dictCode}", Kind: legacyHostRouteDetail},
	{Method: "POST", Path: "/dict/data", Kind: legacyHostRouteCreate},
	{Method: "PUT", Path: "/dict/data/{dictCode}", Kind: legacyHostRouteUpdate},
	{Method: "GET", Path: "/dict/type", Kind: legacyHostRouteList},
	{Method: "GET", Path: "/dict/type/{id}", Kind: legacyHostRouteDetail},
	{Method: "POST", Path: "/dict/type", Kind: legacyHostRouteCreate},
	{Method: "PUT", Path: "/dict/type/{id}", Kind: legacyHostRouteUpdate},
	{Method: "GET", Path: "/config", Kind: legacyHostRouteList},
	{Method: "GET", Path: "/config/{id}", Kind: legacyHostRouteDetail},
	{Method: "POST", Path: "/config", Kind: legacyHostRouteCreate},
	{Method: "PUT", Path: "/config/{id}", Kind: legacyHostRouteUpdate},
}

// registerLegacyRouteInterceptors binds the old-contract response adapter for
// host-owned system routes that cannot be registered again by the plugin.
func registerLegacyRouteInterceptors(global pluginhost.GlobalMiddlewareRegistrar) error {
	if global == nil {
		return nil
	}
	return global.Bind(legacyAPIV1ScopePattern, legacyHostRouteMiddleware)
}

func legacyHostRouteMiddleware(r *ghttp.Request) {
	adapter, ok := legacyMatchHostRouteAdapter(r)
	if !ok {
		r.Middleware.Next()
		return
	}
	legacyApplyHostRouteAliases(r, adapter)
	r.Middleware.Next()
	if !legacyShouldRewriteHostResponse(r) {
		return
	}
	body := r.Response.BufferString()
	if strings.TrimSpace(body) == "" {
		return
	}
	var host legacyHostResponse
	if err := json.Unmarshal([]byte(body), &host); err != nil {
		return
	}
	if !legacyHostResponseSucceeded(host) {
		return
	}
	legacyRewriteHostRouteResponse(r, adapter, host.Data)
}

func legacyShouldRewriteHostResponse(r *ghttp.Request) bool {
	if r == nil || r.Response == nil {
		return false
	}
	status := r.Response.Status
	return status == 0 || status == http.StatusOK
}

func legacyHostResponseSucceeded(host legacyHostResponse) bool {
	return host.Code == legacyHostSuccessCode || host.Code == legacyHTTPStatusOK
}

func legacyRewriteHostRouteResponse(r *ghttp.Request, adapter legacyHostRouteAdapter, data any) {
	r.Response.ClearBuffer()
	switch adapter.Kind {
	case legacyHostRouteList:
		pageIndex, pageSize := legacyHostRoutePage(r)
		writeLegacyHostJSON(r, map[string]any{
			"requestId": legacyHostRequestID(r),
			"code":      legacyHTTPStatusOK,
			"msg":       legacyQuerySuccessMsg,
			"data": map[string]any{
				"count":     legacyHostTotal(data),
				"pageIndex": pageIndex,
				"pageSize":  pageSize,
				"list":      legacyHostList(adapter, data),
			},
		})
	case legacyHostRouteTree:
		writeLegacyHostJSON(r, legacyHostOKEnvelope(legacyHostList(adapter, data), legacyQuerySuccessMsg))
	case legacyHostRouteDetail:
		writeLegacyHostJSON(r, legacyHostOKEnvelope(legacyHostDetail(adapter, data), legacyQuerySuccessMsg))
	case legacyHostRouteCreate:
		writeLegacyHostJSON(r, legacyHostOKEnvelope(legacyHostCreatedID(data), legacyCreateSuccessMsg))
	case legacyHostRouteUpdate:
		writeLegacyHostJSON(r, legacyHostOKEnvelope(legacyRouteIDValue(r, adapter), legacyUpdateSuccessMsg))
	case legacyHostRouteDelete:
		writeLegacyHostJSON(r, legacyHostOKEnvelope(legacyDeletePayload(r, adapter), legacyDeleteSuccessMsg))
	}
}

func writeLegacyHostJSON(r *ghttp.Request, payload map[string]any) {
	if r.Response.Status == 0 {
		r.Response.Status = http.StatusOK
	}
	r.Response.WriteJson(payload)
}

func legacyHostOKEnvelope(data any, msg string) map[string]any {
	return map[string]any{
		"requestId": "",
		"code":      legacyHTTPStatusOK,
		"msg":       msg,
		"data":      data,
	}
}

func legacyHostRequestID(r *ghttp.Request) string {
	for _, key := range []string{"requestId", "RequestId", "traceID", "traceId"} {
		if value := strings.TrimSpace(r.GetCtxVar(key).String()); value != "" {
			return value
		}
	}
	return ""
}

func legacyHostList(adapter legacyHostRouteAdapter, data any) any {
	return legacyMapHostRouteValue(adapter, legacyHostListRaw(data))
}

func legacyHostListRaw(data any) any {
	if data == nil {
		return []map[string]any{}
	}
	if _, ok := data.([]any); ok {
		return data
	}
	record, ok := legacyAsMap(data)
	if !ok {
		return []map[string]any{}
	}
	for _, key := range []string{"list", "items", "rows"} {
		if value, ok := record[key]; ok && value != nil {
			return value
		}
	}
	return []map[string]any{}
}

func legacyHostTotal(data any) int {
	record, ok := legacyAsMap(data)
	if !ok {
		return 0
	}
	for _, key := range []string{"total", "count"} {
		if value, ok := record[key]; ok {
			return gconv.Int(value)
		}
	}
	return 0
}

func legacyHostDetail(adapter legacyHostRouteAdapter, data any) any {
	record, ok := legacyAsMap(data)
	if !ok {
		return map[string]any{}
	}
	if adapter.Path == "/user/profile" {
		return legacyHostProfile(record)
	}
	for _, key := range []string{"item", "detail", "userItem", "configItem", "roleItem", "menuItem", "dictTypeItem", "dictDataItem"} {
		if value, ok := record[key]; ok && value != nil {
			return legacyMapHostRouteValue(adapter, value)
		}
	}
	return legacyMapHostRouteValue(adapter, record)
}

func legacyHostCreatedID(data any) any {
	record, ok := legacyAsMap(data)
	if !ok {
		return 0
	}
	for _, key := range []string{"id", "Id"} {
		if value, ok := record[key]; ok {
			return value
		}
	}
	return 0
}

func legacyRouteIDValue(r *ghttp.Request, adapter legacyHostRouteAdapter) any {
	name := legacyHostRouteParamName(adapter.Path)
	if name == "" {
		return 0
	}
	return gconv.Int64(r.GetRouter(name).String())
}

func legacyDeletePayload(r *ghttp.Request, adapter legacyHostRouteAdapter) any {
	if adapter.Path == "/role" {
		values := r.GetRequest("ids")
		if values.IsEmpty() {
			return nil
		}
		return values.Val()
	}
	return legacyRouteIDValue(r, adapter)
}

func legacyHostRoutePage(r *ghttp.Request) (int, int) {
	pageIndex := gconv.Int(legacyFirstRequestValue(r, "pageIndex", "page_index", "pageNum", "page", "current"))
	pageSize := gconv.Int(legacyFirstRequestValue(r, "pageSize", "page_size", "limit", "size"))
	if pageIndex <= 0 {
		pageIndex = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}
	return pageIndex, pageSize
}

func legacyMatchHostRouteAdapter(r *ghttp.Request) (legacyHostRouteAdapter, bool) {
	if r == nil || r.Request == nil || r.URL == nil {
		return legacyHostRouteAdapter{}, false
	}
	method := strings.ToUpper(r.Method)
	path := legacyTrimAPIV1Path(r.URL.Path)
	for _, adapter := range legacyHostRouteAdapters {
		if adapter.Method != method {
			continue
		}
		if legacyRoutePathMatches(adapter.Path, path) {
			return adapter, true
		}
	}
	return legacyHostRouteAdapter{}, false
}

func legacyTrimAPIV1Path(path string) string {
	path = strings.TrimSpace(path)
	if path == "" {
		return "/"
	}
	if strings.HasPrefix(path, legacyAPIV1Prefix) {
		path = strings.TrimPrefix(path, legacyAPIV1Prefix)
	}
	if path == "" {
		return "/"
	}
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}
	return path
}

func legacyRoutePathMatches(pattern string, path string) bool {
	patternParts := strings.Split(strings.Trim(pattern, "/"), "/")
	pathParts := strings.Split(strings.Trim(path, "/"), "/")
	if len(patternParts) != len(pathParts) {
		return false
	}
	for i, part := range patternParts {
		if strings.HasPrefix(part, "{") && strings.HasSuffix(part, "}") {
			continue
		}
		if part != pathParts[i] {
			return false
		}
	}
	return true
}

func legacyHostRouteParamName(path string) string {
	start := strings.Index(path, "{")
	end := strings.Index(path, "}")
	if start < 0 || end <= start {
		return ""
	}
	return path[start+1 : end]
}

func legacyApplyHostRouteAliases(r *ghttp.Request, adapter legacyHostRouteAdapter) {
	legacyApplyQueryAliases(r, map[string]string{
		"pageIndex":  "pageNum",
		"page_size":  "pageSize",
		"page":       "pageNum",
		"limit":      "pageSize",
		"roleName":   "name",
		"roleKey":    "key",
		"menuName":   "name",
		"title":      "name",
		"dictName":   "name",
		"dictType":   "type",
		"dictLabel":  "label",
		"configName": "name",
		"configKey":  "key",
	})
	legacyApplyJSONAliases(r, adapter)
}

func legacyApplyQueryAliases(r *ghttp.Request, aliases map[string]string) {
	if r == nil || r.URL == nil {
		return
	}
	values := r.URL.Query()
	changed := false
	for from, to := range aliases {
		if values.Get(to) != "" || values.Get(from) == "" {
			continue
		}
		values.Set(to, values.Get(from))
		changed = true
	}
	if changed {
		r.URL.RawQuery = values.Encode()
		r.Request.URL.RawQuery = r.URL.RawQuery
	}
}

func legacyApplyJSONAliases(r *ghttp.Request, adapter legacyHostRouteAdapter) {
	requestMap := r.GetRequestMap()
	if len(requestMap) == 0 {
		return
	}
	switch {
	case strings.HasPrefix(adapter.Path, "/role"):
		legacyAliasMapField(requestMap, "roleName", "name")
		legacyAliasMapField(requestMap, "roleKey", "key")
		legacyAliasMapField(requestMap, "roleSort", "sort")
	case strings.HasPrefix(adapter.Path, "/menu"):
		legacyAliasMapField(requestMap, "menuName", "name")
		legacyAliasMapField(requestMap, "title", "name")
		legacyAliasMapField(requestMap, "permission", "perms")
		legacyAliasMapField(requestMap, "menuType", "type")
		legacyAliasMenuNoCache(requestMap)
	case strings.HasPrefix(adapter.Path, "/dict/data"):
		legacyAliasMapField(requestMap, "dictLabel", "label")
		legacyAliasMapField(requestMap, "dictValue", "value")
		legacyAliasMapField(requestMap, "dictSort", "sort")
		legacyAliasMapField(requestMap, "listClass", "tagStyle")
	case strings.HasPrefix(adapter.Path, "/dict/type"):
		legacyAliasMapField(requestMap, "dictName", "name")
		legacyAliasMapField(requestMap, "dictType", "type")
	case strings.HasPrefix(adapter.Path, "/config"):
		legacyAliasMapField(requestMap, "configName", "name")
		legacyAliasMapField(requestMap, "configKey", "key")
		legacyAliasMapField(requestMap, "configValue", "value")
	}
	if id := legacyRouteIDValue(r, adapter); gconv.Int64(id) > 0 {
		requestMap["id"] = id
	}
	for key, value := range requestMap {
		r.SetParam(key, value)
	}
}

func legacyAliasMapField(data map[string]any, from string, to string) {
	if data == nil || data[to] != nil {
		return
	}
	if value, ok := data[from]; ok {
		data[to] = value
	}
}

func legacyAliasMenuNoCache(data map[string]any) {
	if data == nil || data["isCache"] != nil {
		return
	}
	if value, ok := data["noCache"]; ok {
		if gconv.Bool(value) {
			data["isCache"] = 0
		} else {
			data["isCache"] = 1
		}
	}
}

func legacyFirstRequestValue(r *ghttp.Request, names ...string) string {
	for _, name := range names {
		if value := strings.TrimSpace(r.GetRequest(name).String()); value != "" {
			return value
		}
	}
	return ""
}

type legacyFieldMapping struct {
	Name string
	Keys []string
}

func legacyMapHostRouteValue(adapter legacyHostRouteAdapter, value any) any {
	switch current := value.(type) {
	case []any:
		result := make([]any, 0, len(current))
		for _, item := range current {
			result = append(result, legacyMapHostRouteValue(adapter, item))
		}
		return result
	case []map[string]any:
		result := make([]map[string]any, 0, len(current))
		for _, item := range current {
			result = append(result, legacyMapHostRouteRecord(adapter, item))
		}
		return result
	case map[string]any:
		return legacyMapHostRouteRecord(adapter, current)
	default:
		return value
	}
}

func legacyMapHostRouteRecord(adapter legacyHostRouteAdapter, record map[string]any) map[string]any {
	switch {
	case adapter.Path == "/user/profile":
		return legacyMapSysUser(record)
	case strings.HasPrefix(adapter.Path, "/role"):
		return legacyMapSysRole(record)
	case strings.HasPrefix(adapter.Path, "/menu"):
		return legacyMapSysMenu(record)
	case strings.HasPrefix(adapter.Path, "/dict/data"):
		return legacyMapSysDictData(record)
	case strings.HasPrefix(adapter.Path, "/dict/type"):
		return legacyMapSysDictType(record)
	case strings.HasPrefix(adapter.Path, "/config"):
		return legacyMapSysConfig(record)
	default:
		return record
	}
}

func legacyHostProfile(data map[string]any) map[string]any {
	user := data
	if value, ok := legacyFirstMapValue(data, "user", "userInfo", "profile"); ok {
		if nested, nestedOK := legacyAsMap(value); nestedOK {
			user = nested
		}
	}
	result := map[string]any{
		"user":    legacyMapSysUser(user),
		"roles":   legacyMapRecords(firstExistingMapValue(data, "roles", "roleList"), legacyMapSysRole),
		"posts":   legacyMapRecords(firstExistingMapValue(data, "posts", "postList"), legacyMapSysPost),
		"roleIds": legacyExistingOrEmpty(data, "roleIds", "role_ids"),
		"postIds": legacyExistingOrEmpty(data, "postIds", "post_ids"),
	}
	return result
}

func legacyMapSysRole(record map[string]any) map[string]any {
	result := legacyMapFields(record, []legacyFieldMapping{
		{Name: "roleId", Keys: []string{"roleId", "role_id", "id"}},
		{Name: "roleName", Keys: []string{"roleName", "role_name", "name"}},
		{Name: "status", Keys: []string{"status"}},
		{Name: "roleKey", Keys: []string{"roleKey", "role_key", "key", "code"}},
		{Name: "roleSort", Keys: []string{"roleSort", "role_sort", "sort", "order"}},
		{Name: "flag", Keys: []string{"flag"}},
		{Name: "remark", Keys: []string{"remark"}},
		{Name: "admin", Keys: []string{"admin", "isAdmin", "is_admin"}},
		{Name: "dataScope", Keys: []string{"dataScope", "data_scope"}},
		{Name: "params", Keys: []string{"params"}},
		{Name: "menuIds", Keys: []string{"menuIds", "menu_ids"}},
		{Name: "deptIds", Keys: []string{"deptIds", "dept_ids"}},
		{Name: "createdAt", Keys: []string{"createdAt", "created_at"}},
		{Name: "updatedAt", Keys: []string{"updatedAt", "updated_at"}},
		{Name: "deletedAt", Keys: []string{"deletedAt", "deleted_at"}},
		{Name: "createBy", Keys: []string{"createBy", "create_by"}},
		{Name: "updateBy", Keys: []string{"updateBy", "update_by"}},
	})
	if value, ok := legacyFirstMapValue(record, "sysMenu", "menus"); ok {
		result["sysMenu"] = legacyMapRecords(value, legacyMapSysMenu)
	}
	if value, ok := legacyFirstMapValue(record, "sysDept", "depts"); ok {
		result["sysDept"] = legacyMapRecords(value, legacyMapSysDept)
	}
	return result
}

func legacyMapSysMenu(record map[string]any) map[string]any {
	result := legacyMapFields(record, []legacyFieldMapping{
		{Name: "menuId", Keys: []string{"menuId", "menu_id", "id"}},
		{Name: "menuName", Keys: []string{"menuName", "menu_name", "name"}},
		{Name: "title", Keys: []string{"title", "name"}},
		{Name: "icon", Keys: []string{"icon"}},
		{Name: "path", Keys: []string{"path"}},
		{Name: "paths", Keys: []string{"paths"}},
		{Name: "menuType", Keys: []string{"menuType", "menu_type", "type"}},
		{Name: "action", Keys: []string{"action"}},
		{Name: "permission", Keys: []string{"permission", "perms"}},
		{Name: "parentId", Keys: []string{"parentId", "parent_id"}},
		{Name: "noCache", Keys: []string{"noCache", "no_cache"}},
		{Name: "breadcrumb", Keys: []string{"breadcrumb"}},
		{Name: "component", Keys: []string{"component"}},
		{Name: "sort", Keys: []string{"sort", "order"}},
		{Name: "visible", Keys: []string{"visible"}},
		{Name: "isFrame", Keys: []string{"isFrame", "is_frame"}},
		{Name: "apis", Keys: []string{"apis"}},
		{Name: "dataScope", Keys: []string{"dataScope", "data_scope"}},
		{Name: "params", Keys: []string{"params"}},
		{Name: "roleId", Keys: []string{"roleId", "role_id"}},
		{Name: "is_select", Keys: []string{"is_select", "isSelect"}},
		{Name: "createdAt", Keys: []string{"createdAt", "created_at"}},
		{Name: "updatedAt", Keys: []string{"updatedAt", "updated_at"}},
		{Name: "deletedAt", Keys: []string{"deletedAt", "deleted_at"}},
		{Name: "createBy", Keys: []string{"createBy", "create_by"}},
		{Name: "updateBy", Keys: []string{"updateBy", "update_by"}},
	})
	if value, ok := legacyFirstMapValue(record, "sysApi", "apisDetail", "apisData"); ok {
		result["sysApi"] = value
	}
	if _, ok := result["noCache"]; !ok {
		if value, valueOK := legacyFirstMapValue(record, "isCache"); valueOK {
			result["noCache"] = !gconv.Bool(value)
		}
	}
	if value, ok := legacyFirstMapValue(record, "children"); ok {
		result["children"] = legacyMapRecords(value, legacyMapSysMenu)
	}
	return result
}

func legacyMapSysDictData(record map[string]any) map[string]any {
	return legacyMapFields(record, []legacyFieldMapping{
		{Name: "dictCode", Keys: []string{"dictCode", "dict_code", "id"}},
		{Name: "dictSort", Keys: []string{"dictSort", "dict_sort", "sort", "order"}},
		{Name: "dictLabel", Keys: []string{"dictLabel", "dict_label", "label", "name"}},
		{Name: "dictValue", Keys: []string{"dictValue", "dict_value", "value"}},
		{Name: "dictType", Keys: []string{"dictType", "dict_type", "type"}},
		{Name: "cssClass", Keys: []string{"cssClass", "css_class"}},
		{Name: "listClass", Keys: []string{"listClass", "list_class", "tagStyle"}},
		{Name: "isDefault", Keys: []string{"isDefault", "is_default"}},
		{Name: "status", Keys: []string{"status"}},
		{Name: "default", Keys: []string{"default"}},
		{Name: "remark", Keys: []string{"remark"}},
		{Name: "createdAt", Keys: []string{"createdAt", "created_at"}},
		{Name: "updatedAt", Keys: []string{"updatedAt", "updated_at"}},
		{Name: "deletedAt", Keys: []string{"deletedAt", "deleted_at"}},
		{Name: "createBy", Keys: []string{"createBy", "create_by"}},
		{Name: "updateBy", Keys: []string{"updateBy", "update_by"}},
	})
}

func legacyMapSysDictType(record map[string]any) map[string]any {
	result := legacyMapFields(record, []legacyFieldMapping{
		{Name: "id", Keys: []string{"id", "dictId", "dict_id"}},
		{Name: "dictName", Keys: []string{"dictName", "dict_name", "name"}},
		{Name: "dictType", Keys: []string{"dictType", "dict_type", "type", "key", "code"}},
		{Name: "status", Keys: []string{"status"}},
		{Name: "remark", Keys: []string{"remark"}},
		{Name: "createdAt", Keys: []string{"createdAt", "created_at"}},
		{Name: "updatedAt", Keys: []string{"updatedAt", "updated_at"}},
		{Name: "deletedAt", Keys: []string{"deletedAt", "deleted_at"}},
		{Name: "createBy", Keys: []string{"createBy", "create_by"}},
		{Name: "updateBy", Keys: []string{"updateBy", "update_by"}},
	})
	if value, ok := legacyFirstMapValue(record, "dictId", "dict_id", "id"); ok {
		result["dictId"] = value
	}
	return result
}

func legacyMapSysConfig(record map[string]any) map[string]any {
	result := legacyMapFields(record, []legacyFieldMapping{
		{Name: "id", Keys: []string{"id", "configId", "config_id"}},
		{Name: "configName", Keys: []string{"configName", "config_name", "name"}},
		{Name: "configKey", Keys: []string{"configKey", "config_key", "key"}},
		{Name: "configValue", Keys: []string{"configValue", "config_value", "value"}},
		{Name: "configType", Keys: []string{"configType", "config_type", "type"}},
		{Name: "isFrontend", Keys: []string{"isFrontend", "is_frontend"}},
		{Name: "remark", Keys: []string{"remark"}},
		{Name: "createdAt", Keys: []string{"createdAt", "created_at"}},
		{Name: "updatedAt", Keys: []string{"updatedAt", "updated_at"}},
		{Name: "deletedAt", Keys: []string{"deletedAt", "deleted_at"}},
		{Name: "createBy", Keys: []string{"createBy", "create_by"}},
		{Name: "updateBy", Keys: []string{"updateBy", "update_by"}},
	})
	if value, ok := legacyFirstMapValue(record, "configId", "config_id", "id"); ok {
		result["configId"] = value
	}
	return result
}

func legacyMapSysUser(record map[string]any) map[string]any {
	result := legacyMapFields(record, []legacyFieldMapping{
		{Name: "userId", Keys: []string{"userId", "user_id", "id"}},
		{Name: "username", Keys: []string{"username"}},
		{Name: "nickName", Keys: []string{"nickName", "nick_name", "nickname", "name"}},
		{Name: "phone", Keys: []string{"phone", "mobile"}},
		{Name: "roleId", Keys: []string{"roleId", "role_id"}},
		{Name: "avatar", Keys: []string{"avatar"}},
		{Name: "sex", Keys: []string{"sex", "gender"}},
		{Name: "email", Keys: []string{"email"}},
		{Name: "deptId", Keys: []string{"deptId", "dept_id"}},
		{Name: "postId", Keys: []string{"postId", "post_id"}},
		{Name: "remark", Keys: []string{"remark"}},
		{Name: "status", Keys: []string{"status"}},
		{Name: "deptIds", Keys: []string{"deptIds", "dept_ids"}},
		{Name: "postIds", Keys: []string{"postIds", "post_ids"}},
		{Name: "roleIds", Keys: []string{"roleIds", "role_ids"}},
		{Name: "createdAt", Keys: []string{"createdAt", "created_at"}},
		{Name: "updatedAt", Keys: []string{"updatedAt", "updated_at"}},
		{Name: "deletedAt", Keys: []string{"deletedAt", "deleted_at"}},
		{Name: "createBy", Keys: []string{"createBy", "create_by"}},
		{Name: "updateBy", Keys: []string{"updateBy", "update_by"}},
	})
	if value, ok := legacyFirstMapValue(record, "dept"); ok {
		if nested, nestedOK := legacyAsMap(value); nestedOK {
			result["dept"] = legacyMapSysDept(nested)
		} else {
			result["dept"] = value
		}
	}
	return result
}

func legacyMapSysDept(record map[string]any) map[string]any {
	result := legacyMapFields(record, []legacyFieldMapping{
		{Name: "deptId", Keys: []string{"deptId", "dept_id", "id"}},
		{Name: "parentId", Keys: []string{"parentId", "parent_id"}},
		{Name: "deptPath", Keys: []string{"deptPath", "dept_path"}},
		{Name: "deptName", Keys: []string{"deptName", "dept_name", "name"}},
		{Name: "sort", Keys: []string{"sort", "order"}},
		{Name: "leader", Keys: []string{"leader"}},
		{Name: "phone", Keys: []string{"phone"}},
		{Name: "email", Keys: []string{"email"}},
		{Name: "status", Keys: []string{"status"}},
		{Name: "dataScope", Keys: []string{"dataScope", "data_scope"}},
		{Name: "params", Keys: []string{"params"}},
		{Name: "createdAt", Keys: []string{"createdAt", "created_at"}},
		{Name: "updatedAt", Keys: []string{"updatedAt", "updated_at"}},
		{Name: "deletedAt", Keys: []string{"deletedAt", "deleted_at"}},
		{Name: "createBy", Keys: []string{"createBy", "create_by"}},
		{Name: "updateBy", Keys: []string{"updateBy", "update_by"}},
	})
	if value, ok := legacyFirstMapValue(record, "children"); ok {
		result["children"] = legacyMapRecords(value, legacyMapSysDept)
	}
	return result
}

func legacyMapSysPost(record map[string]any) map[string]any {
	return legacyMapFields(record, []legacyFieldMapping{
		{Name: "postId", Keys: []string{"postId", "post_id", "id"}},
		{Name: "postName", Keys: []string{"postName", "post_name", "name"}},
		{Name: "postCode", Keys: []string{"postCode", "post_code", "code", "key"}},
		{Name: "sort", Keys: []string{"sort", "order"}},
		{Name: "status", Keys: []string{"status"}},
		{Name: "remark", Keys: []string{"remark"}},
		{Name: "dataScope", Keys: []string{"dataScope", "data_scope"}},
		{Name: "params", Keys: []string{"params"}},
		{Name: "createdAt", Keys: []string{"createdAt", "created_at"}},
		{Name: "updatedAt", Keys: []string{"updatedAt", "updated_at"}},
		{Name: "deletedAt", Keys: []string{"deletedAt", "deleted_at"}},
		{Name: "createBy", Keys: []string{"createBy", "create_by"}},
		{Name: "updateBy", Keys: []string{"updateBy", "update_by"}},
	})
}

func legacyMapFields(record map[string]any, mappings []legacyFieldMapping) map[string]any {
	result := make(map[string]any, len(mappings))
	for _, mapping := range mappings {
		if value, ok := legacyFirstMapValue(record, mapping.Keys...); ok {
			result[mapping.Name] = value
		}
	}
	return result
}

func legacyFirstMapValue(record map[string]any, keys ...string) (any, bool) {
	if record == nil {
		return nil, false
	}
	for _, key := range keys {
		if value, ok := record[key]; ok && value != nil {
			return value, true
		}
	}
	for _, key := range keys {
		for currentKey, value := range record {
			if strings.EqualFold(currentKey, key) && value != nil {
				return value, true
			}
		}
	}
	return nil, false
}

func firstExistingMapValue(record map[string]any, keys ...string) any {
	if value, ok := legacyFirstMapValue(record, keys...); ok {
		return value
	}
	return nil
}

func legacyExistingOrEmpty(record map[string]any, keys ...string) any {
	if value, ok := legacyFirstMapValue(record, keys...); ok {
		return value
	}
	return []int{}
}

func legacyMapRecords(value any, mapper func(map[string]any) map[string]any) any {
	switch current := value.(type) {
	case []any:
		result := make([]any, 0, len(current))
		for _, item := range current {
			if record, ok := legacyAsMap(item); ok {
				result = append(result, mapper(record))
			} else {
				result = append(result, item)
			}
		}
		return result
	case []map[string]any:
		result := make([]map[string]any, 0, len(current))
		for _, item := range current {
			result = append(result, mapper(item))
		}
		return result
	case map[string]any:
		return mapper(current)
	default:
		return []map[string]any{}
	}
}

func legacyAsMap(value any) (map[string]any, bool) {
	switch current := value.(type) {
	case map[string]any:
		return current, true
	default:
		return nil, false
	}
}
