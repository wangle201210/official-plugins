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
	Code    int            `json:"code"`
	Message string         `json:"message"`
	Msg     string         `json:"msg"`
	Data    map[string]any `json:"data"`
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

func legacyRewriteHostRouteResponse(r *ghttp.Request, adapter legacyHostRouteAdapter, data map[string]any) {
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
				"list":      legacyHostList(data),
			},
		})
	case legacyHostRouteTree:
		writeLegacyHostJSON(r, legacyHostOKEnvelope(legacyHostList(data), legacyQuerySuccessMsg))
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

func legacyHostList(data map[string]any) any {
	if data == nil {
		return []map[string]any{}
	}
	for _, key := range []string{"list", "items", "rows"} {
		if value, ok := data[key]; ok && value != nil {
			return value
		}
	}
	return []map[string]any{}
}

func legacyHostTotal(data map[string]any) int {
	if data == nil {
		return 0
	}
	for _, key := range []string{"total", "count"} {
		if value, ok := data[key]; ok {
			return gconv.Int(value)
		}
	}
	return 0
}

func legacyHostDetail(adapter legacyHostRouteAdapter, data map[string]any) any {
	if data == nil {
		return map[string]any{}
	}
	if adapter.Path == "/user/profile" {
		return map[string]any{
			"user":    data,
			"roles":   []map[string]any{},
			"posts":   []map[string]any{},
			"roleIds": []int{},
			"postIds": []int{},
		}
	}
	for _, key := range []string{"item", "detail", "userItem", "configItem", "roleItem", "menuItem", "dictTypeItem", "dictDataItem"} {
		if value, ok := data[key]; ok && value != nil {
			return value
		}
	}
	return data
}

func legacyHostCreatedID(data map[string]any) any {
	if data == nil {
		return 0
	}
	for _, key := range []string{"id", "Id"} {
		if value, ok := data[key]; ok {
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
		legacyAliasMapField(requestMap, "noCache", "isCache")
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

func legacyFirstRequestValue(r *ghttp.Request, names ...string) string {
	for _, name := range names {
		if value := strings.TrimSpace(r.GetRequest(name).String()); value != "" {
			return value
		}
	}
	return ""
}
