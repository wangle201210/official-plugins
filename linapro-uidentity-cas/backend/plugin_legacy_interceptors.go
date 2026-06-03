// This file dispatches old uidentity/admin routes that collide with LinaPro host
// system-management routes. The plugin cannot register duplicate static routes,
// so it uses guarded global middleware to selectively reuse the already
// published host request middlewares and then serves the old contract from plugin-owned
// compatibility tables.

package backend

import (
	"strings"

	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/gogf/gf/v2/util/gconv"

	"lina-core/pkg/plugin/pluginhost"
	uidentitycontroller "lina-plugin-linapro-uidentity-cas/backend/internal/controller/uidentity"
)

const (
	legacyAPIV1Prefix       = "/api/v1"
	legacyAPIV1ScopePattern = "/api/v1/*"
)

type legacyHostRouteAdapter struct {
	Method  string
	Path    string
	Handler ghttp.HandlerFunc
}

// registerLegacyRouteInterceptors binds the old-contract dispatcher for
// host-owned system route paths that cannot be registered again by the plugin.
func registerLegacyRouteInterceptors(
	global pluginhost.GlobalMiddlewareRegistrar,
	middlewares pluginhost.RouteMiddlewares,
	legacyController *uidentitycontroller.LegacyController,
) error {
	if global == nil {
		return nil
	}
	if middlewares == nil || middlewares.Ctx() == nil || middlewares.Auth() == nil || middlewares.Tenancy() == nil {
		return gerror.New("linapro-uidentity-cas legacy route interceptors require host ctx, auth, and tenancy middlewares")
	}
	adapters := legacyHostRouteAdapters(legacyController)
	for _, middleware := range []pluginhost.RouteMiddleware{
		middlewares.Ctx(),
		middlewares.Auth(),
		middlewares.Tenancy(),
	} {
		next := middleware
		if err := global.Bind(legacyAPIV1ScopePattern, func(r *ghttp.Request) {
			legacyHostRouteGuardedMiddleware(r, adapters, next)
		}); err != nil {
			return err
		}
	}
	return global.Bind(legacyAPIV1ScopePattern, func(r *ghttp.Request) {
		legacyHostRouteDispatchMiddleware(r, adapters)
	})
}

func legacyHostRouteAdapters(controller *uidentitycontroller.LegacyController) []legacyHostRouteAdapter {
	if controller == nil {
		return nil
	}
	return []legacyHostRouteAdapter{
		{Method: "GET", Path: "/user/profile", Handler: controller.LegacySystemProfile},
		{Method: "GET", Path: "/role", Handler: controller.LegacySystemList("roles")},
		{Method: "GET", Path: "/role/{id}", Handler: controller.LegacySystemGet("roles")},
		{Method: "POST", Path: "/role", Handler: controller.LegacySystemCreate("roles")},
		{Method: "PUT", Path: "/role/{id}", Handler: controller.LegacySystemUpdate("roles")},
		{Method: "DELETE", Path: "/role", Handler: controller.LegacySystemDelete("roles")},
		{Method: "GET", Path: "/menu", Handler: controller.LegacyMenuList},
		{Method: "GET", Path: "/menu/{id}", Handler: controller.LegacySystemGet("menus")},
		{Method: "POST", Path: "/menu", Handler: controller.LegacySystemCreate("menus")},
		{Method: "PUT", Path: "/menu/{id}", Handler: controller.LegacySystemUpdate("menus")},
		{Method: "GET", Path: "/dict/data", Handler: controller.LegacySystemList("dict-data")},
		{Method: "GET", Path: "/dict/data/{dictCode}", Handler: controller.LegacySystemGet("dict-data")},
		{Method: "POST", Path: "/dict/data", Handler: controller.LegacySystemCreate("dict-data")},
		{Method: "PUT", Path: "/dict/data/{dictCode}", Handler: controller.LegacySystemUpdate("dict-data")},
		{Method: "GET", Path: "/dict/type", Handler: controller.LegacySystemList("dict-types")},
		{Method: "GET", Path: "/dict/type/{id}", Handler: controller.LegacySystemGet("dict-types")},
		{Method: "POST", Path: "/dict/type", Handler: controller.LegacySystemCreate("dict-types")},
		{Method: "PUT", Path: "/dict/type/{id}", Handler: controller.LegacySystemUpdate("dict-types")},
		{Method: "GET", Path: "/config", Handler: controller.LegacySystemList("configs")},
		{Method: "GET", Path: "/config/{id}", Handler: controller.LegacySystemGet("configs")},
		{Method: "POST", Path: "/config", Handler: controller.LegacySystemCreate("configs")},
		{Method: "PUT", Path: "/config/{id}", Handler: controller.LegacySystemUpdate("configs")},
	}
}

func legacyHostRouteGuardedMiddleware(
	r *ghttp.Request,
	adapters []legacyHostRouteAdapter,
	next pluginhost.RouteMiddleware,
) {
	_, ok := legacyMatchHostRouteAdapter(r, adapters)
	if !ok {
		r.Middleware.Next()
		return
	}
	next(r)
}

func legacyHostRouteDispatchMiddleware(
	r *ghttp.Request,
	adapters []legacyHostRouteAdapter,
) {
	adapter, ok := legacyMatchHostRouteAdapter(r, adapters)
	if !ok {
		r.Middleware.Next()
		return
	}
	legacyApplyHostRouteAliases(r, adapter)
	if adapter.Handler != nil {
		adapter.Handler(r)
	}
}

func legacyMatchHostRouteAdapter(r *ghttp.Request, adapters []legacyHostRouteAdapter) (legacyHostRouteAdapter, bool) {
	if r == nil || r.Request == nil || r.URL == nil {
		return legacyHostRouteAdapter{}, false
	}
	method := strings.ToUpper(r.Method)
	path := legacyTrimAPIV1Path(r.URL.Path)
	for _, adapter := range adapters {
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

func legacyApplyHostRouteAliases(r *ghttp.Request, adapter legacyHostRouteAdapter) {
	legacyApplyRoutePathParams(r, adapter)
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

func legacyApplyRoutePathParams(r *ghttp.Request, adapter legacyHostRouteAdapter) {
	if r == nil || r.URL == nil {
		return
	}
	params := legacyRoutePathParams(adapter.Path, legacyTrimAPIV1Path(r.URL.Path))
	for key, value := range params {
		r.SetParam(key, value)
	}
}

func legacyRoutePathParams(pattern string, path string) map[string]string {
	patternParts := strings.Split(strings.Trim(pattern, "/"), "/")
	pathParts := strings.Split(strings.Trim(path, "/"), "/")
	if len(patternParts) != len(pathParts) {
		return nil
	}
	result := make(map[string]string)
	for i, part := range patternParts {
		if !strings.HasPrefix(part, "{") || !strings.HasSuffix(part, "}") {
			continue
		}
		name := strings.TrimSuffix(strings.TrimPrefix(part, "{"), "}")
		if strings.TrimSpace(name) == "" {
			continue
		}
		result[name] = pathParts[i]
	}
	return result
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
	if id := legacyHostRouteIDParam(r, adapter); id > 0 {
		requestMap["id"] = id
	}
	for key, value := range requestMap {
		r.SetParam(key, value)
	}
}

func legacyHostRouteIDParam(r *ghttp.Request, adapter legacyHostRouteAdapter) int64 {
	name := legacyHostRouteParamName(adapter.Path)
	if name == "" {
		return 0
	}
	return gconv.Int64(r.GetRequest(name).String())
}

func legacyHostRouteParamName(path string) string {
	start := strings.Index(path, "{")
	end := strings.Index(path, "}")
	if start < 0 || end <= start {
		return ""
	}
	return path[start+1 : end]
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
