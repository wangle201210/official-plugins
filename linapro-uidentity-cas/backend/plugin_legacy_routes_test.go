// This file verifies that the compatibility route inventory matches the old
// uidentity/admin router source inventory method-by-method and path-by-path.

package backend

import (
	"strings"
	"testing"
)

func TestLegacyRouteSpecCoversOldAdminRouterInventory(t *testing.T) {
	coverage := legacyRouteCoverageSetForTest()
	for _, item := range oldAdminRouterInventoryForTest() {
		key := item.Method + " " + item.Path
		if _, ok := coverage[key]; !ok {
			t.Fatalf("old uidentity/admin route missing: %s", key)
		}
	}
}

func TestLegacyRouteSpecExcludesNonOldAPIv1Contracts(t *testing.T) {
	publicRoutes := legacyRouteSpecsForTest(legacyPublicRoutes)
	for _, item := range []legacyRouteSpec{
		{Method: "ALL", Path: "/sso/serviceValidate"},
		{Method: "ALL", Path: "/sso/proxyValidate"},
		{Method: "GET", Path: "/sso/login"},
		{Method: "ALL", Path: "/sso/logout"},
		{Method: "POST", Path: "/ssologin/getToken"},
		{Method: "POST", Path: "/user/changeWechatCallBack"},
	} {
		key := item.Method + " " + item.Path
		if _, ok := publicRoutes[key]; ok {
			t.Fatalf("legacy API v1 route should not be registered: %s", key)
		}
	}
	protectedRoutes := legacyRouteSpecsForTest(legacyProtectedRoutes)
	if _, ok := protectedRoutes["POST /ldap/sync"]; ok {
		t.Fatalf("legacy API v1 route should not be registered: POST /ldap/sync")
	}
}

func TestLegacyResourceRouteNamesMatchOldAdminPaths(t *testing.T) {
	got := map[string]string{}
	for _, route := range legacyResourceRoutes {
		got[route.Path] = route.Resource
	}
	want := map[string]string{
		"/account":               "accounts",
		"/account-details":       "account-details",
		"/account-unit":          "account-unit",
		"/account-app-role":      "account-app-role",
		"/account-app-blacklist": "account-app-blacklist",
		"/account-change-log":    "account-change-log",
		"/units":                 "units",
		"/groups":                "groups",
		"/containers":            "containers",
		"/applications":          "applications",
		"/group-app-blacklist":   "group-app-blacklist",
		"/pass-ruler":            "pass-ruler",
		"/sms":                   "sms",
		"/cas-login-log":         "cas-login-logs-legacy",
		"/oauth-log":             "oauth-log",
		"/oauth-token":           "oauth-token",
	}
	for path, resource := range want {
		if got[path] != resource {
			t.Fatalf("legacy resource path %s maps to %q, want %q", path, got[path], resource)
		}
	}
}

func TestLegacyHostCoveredOldRoutesAreExactOldPaths(t *testing.T) {
	oldRoutes := legacyRouteSpecsForTest(oldAdminRouterInventoryForTest())
	for _, route := range legacyHostCoveredOldRoutes {
		key := route.Method + " " + route.Path
		if _, ok := oldRoutes[key]; !ok {
			t.Fatalf("host-covered route is not present in old router inventory: %s", key)
		}
	}
}

func legacyRouteCoverageSetForTest() map[string]struct{} {
	routes := append([]legacyRouteSpec{}, allLegacyRouteSpecs()...)
	routes = append(routes, legacyHostCoveredOldRoutes...)
	return legacyRouteSpecsForTest(routes)
}

func legacyRouteSpecsForTest(routes []legacyRouteSpec) map[string]struct{} {
	got := make(map[string]struct{}, len(routes))
	for _, route := range routes {
		got[strings.ToUpper(route.Method)+" "+route.Path] = struct{}{}
	}
	return got
}

func oldAdminRouterInventoryForTest() []legacyRouteSpec {
	routes := []legacyRouteSpec{
		{Method: "GET", Path: "/"},
		{Method: "GET", Path: "/info"},
		{Method: "GET", Path: "/swagger/admin/*any"},
		{Method: "GET", Path: "/ws/{id}/{channel}"},
		{Method: "GET", Path: "/wslogout/{id}/{channel}"},
		{Method: "GET", Path: "/static/*filepath"},
		{Method: "HEAD", Path: "/static/*filepath"},
		{Method: "GET", Path: "/form-generator/*filepath"},
		{Method: "HEAD", Path: "/form-generator/*filepath"},
		{Method: "GET", Path: "/logs/*filepath"},
		{Method: "HEAD", Path: "/logs/*filepath"},
		{Method: "ALL", Path: "/wechat/callback"},
		{Method: "ALL", Path: "/MP_verify_5osfGmdqMLsyyzYp.txt"},
		{Method: "ALL", Path: "/EcEOCIhE9w.txt"},
		{Method: "ALL", Path: "/sso/serviceValidate"},
		{Method: "ALL", Path: "/sso/proxyValidate"},
		{Method: "GET", Path: "/sso/login"},
		{Method: "ALL", Path: "/sso/logout"},
		{Method: "POST", Path: "/ssologin/getToken"},
		{Method: "POST", Path: "/login"},
		{Method: "GET", Path: "/refresh_token"},
		{Method: "POST", Path: "/logout"},
		{Method: "GET", Path: "/captcha"},
		{Method: "GET", Path: "/health"},
		{Method: "GET", Path: "/metrics"},
		{Method: "GET", Path: "/roleMenuTreeselect/{roleId}"},
		{Method: "GET", Path: "/roleDeptTreeselect/{roleId}"},
		{Method: "GET", Path: "/cas/login"},
		{Method: "POST", Path: "/cas/login"},
		{Method: "POST", Path: "/cas/loginByPhone"},
		{Method: "DELETE", Path: "/cas/tickets/{ticket}"},
		{Method: "GET", Path: "/cas/proxyValidate"},
		{Method: "POST", Path: "/cas/getLoginQr"},
		{Method: "GET", Path: "/cas/loginByQr"},
		{Method: "POST", Path: "/cas/getCasLoginQrRes"},
		{Method: "GET", Path: "/cas/loginByUnionID"},
		{Method: "GET", Path: "/cas-login/index"},
		{Method: "POST", Path: "/token/get"},
		{Method: "GET", Path: "/token/getUserInfoByToken"},
		{Method: "GET", Path: "/wechat/login"},
		{Method: "GET", Path: "/wechat/loginCallback"},
		{Method: "POST", Path: "/activate/baseInfo"},
		{Method: "POST", Path: "/activate/face"},
		{Method: "POST", Path: "/activate/password"},
		{Method: "POST", Path: "/activate/phone"},
		{Method: "POST", Path: "/activate/wechatQr"},
		{Method: "GET", Path: "/activate/wechatScan"},
		{Method: "POST", Path: "/activate/state"},
		{Method: "POST", Path: "/user/getByUnionID"},
		{Method: "GET", Path: "/user/bindUnionIDCallBack"},
		{Method: "POST", Path: "/user/bindUnionID"},
		{Method: "POST", Path: "/user/changePassword"},
		{Method: "POST", Path: "/user/changePhone"},
		{Method: "POST", Path: "/user/changeEmail"},
		{Method: "POST", Path: "/user/changeQQ"},
		{Method: "POST", Path: "/user/unbindWechat"},
		{Method: "GET", Path: "/user/getUserCasLoginLog"},
		{Method: "POST", Path: "/user/getUserInfo"},
		{Method: "GET", Path: "/user/accountAppList"},
		{Method: "GET", Path: "/user/accountAppRole"},
		{Method: "POST", Path: "/user/accountAppRole"},
		{Method: "POST", Path: "/user/accountAppRoleUpdate"},
		{Method: "POST", Path: "/user/changeWechatQr"},
		{Method: "POST", Path: "/user/changeWechatState"},
		{Method: "GET", Path: "/oauth/login"},
		{Method: "POST", Path: "/oauth/login"},
		{Method: "ALL", Path: "/oauth/auth"},
		{Method: "ALL", Path: "/oauth/authorize"},
		{Method: "POST", Path: "/oauth/token"},
		{Method: "ALL", Path: "/oauth/test"},
		{Method: "POST", Path: "/sms/send"},
		{Method: "GET", Path: "/stat/get"},
		{Method: "GET", Path: "/config/cas"},
		{Method: "GET", Path: "/config/ldap"},
		{Method: "GET", Path: "/config/oauth"},
		{Method: "GET", Path: "/config/token"},
		{Method: "GET", Path: "/server-monitor"},
		{Method: "POST", Path: "/public/uploadFile"},
		{Method: "GET", Path: "/log/watch"},
		{Method: "GET", Path: "/gen/preview/{tableId}"},
		{Method: "GET", Path: "/gen/toproject/{tableId}"},
		{Method: "GET", Path: "/gen/apitofile/{tableId}"},
		{Method: "GET", Path: "/gen/todb/{tableId}"},
		{Method: "GET", Path: "/gen/tabletree"},
		{Method: "GET", Path: "/db/tables/page"},
		{Method: "GET", Path: "/db/columns/page"},
		{Method: "GET", Path: "/sys/tables/page"},
		{Method: "POST", Path: "/sys/tables/info"},
		{Method: "PUT", Path: "/sys/tables/info"},
		{Method: "DELETE", Path: "/sys/tables/info/{tableId}"},
		{Method: "GET", Path: "/sys/tables/info/{tableId}"},
		{Method: "GET", Path: "/sys/tables/info"},
		{Method: "GET", Path: "/sys-user"},
		{Method: "GET", Path: "/sys-user/{id}"},
		{Method: "POST", Path: "/sys-user"},
		{Method: "PUT", Path: "/sys-user"},
		{Method: "DELETE", Path: "/sys-user"},
		{Method: "GET", Path: "/user/profile"},
		{Method: "POST", Path: "/user/avatar"},
		{Method: "PUT", Path: "/user/pwd/set"},
		{Method: "PUT", Path: "/user/pwd/reset"},
		{Method: "PUT", Path: "/user/status"},
		{Method: "GET", Path: "/getinfo"},
		{Method: "GET", Path: "/sys-login-log"},
		{Method: "GET", Path: "/sys-login-log/{id}"},
		{Method: "DELETE", Path: "/sys-login-log"},
		{Method: "GET", Path: "/sys-opera-log"},
		{Method: "GET", Path: "/sys-opera-log/{id}"},
		{Method: "DELETE", Path: "/sys-opera-log"},
		{Method: "GET", Path: "/sys-api"},
		{Method: "GET", Path: "/sys-api/{id}"},
		{Method: "PUT", Path: "/sys-api/{id}"},
		{Method: "GET", Path: "/role"},
		{Method: "GET", Path: "/role/{id}"},
		{Method: "POST", Path: "/role"},
		{Method: "PUT", Path: "/role/{id}"},
		{Method: "DELETE", Path: "/role"},
		{Method: "PUT", Path: "/role-status"},
		{Method: "PUT", Path: "/roledatascope"},
		{Method: "GET", Path: "/menu"},
		{Method: "GET", Path: "/menu/{id}"},
		{Method: "POST", Path: "/menu"},
		{Method: "PUT", Path: "/menu/{id}"},
		{Method: "DELETE", Path: "/menu"},
		{Method: "GET", Path: "/menurole"},
		{Method: "GET", Path: "/dept"},
		{Method: "GET", Path: "/dept/{id}"},
		{Method: "POST", Path: "/dept"},
		{Method: "PUT", Path: "/dept/{id}"},
		{Method: "DELETE", Path: "/dept"},
		{Method: "GET", Path: "/deptTree"},
		{Method: "GET", Path: "/post"},
		{Method: "GET", Path: "/post/{id}"},
		{Method: "POST", Path: "/post"},
		{Method: "PUT", Path: "/post/{id}"},
		{Method: "DELETE", Path: "/post"},
		{Method: "GET", Path: "/dict/data"},
		{Method: "GET", Path: "/dict/data/{dictCode}"},
		{Method: "POST", Path: "/dict/data"},
		{Method: "PUT", Path: "/dict/data/{dictCode}"},
		{Method: "DELETE", Path: "/dict/data"},
		{Method: "GET", Path: "/dict/type-option-select"},
		{Method: "GET", Path: "/dict/type"},
		{Method: "GET", Path: "/dict/type/{id}"},
		{Method: "POST", Path: "/dict/type"},
		{Method: "PUT", Path: "/dict/type/{id}"},
		{Method: "DELETE", Path: "/dict/type"},
		{Method: "GET", Path: "/dict-data/option-select"},
		{Method: "GET", Path: "/config"},
		{Method: "GET", Path: "/config/{id}"},
		{Method: "POST", Path: "/config"},
		{Method: "PUT", Path: "/config/{id}"},
		{Method: "DELETE", Path: "/config"},
		{Method: "GET", Path: "/configKey/{configKey}"},
		{Method: "GET", Path: "/app-config"},
		{Method: "PUT", Path: "/set-config"},
		{Method: "GET", Path: "/set-config"},
		{Method: "GET", Path: "/sysjob"},
		{Method: "GET", Path: "/sysjob/{id}"},
		{Method: "POST", Path: "/sysjob"},
		{Method: "PUT", Path: "/sysjob"},
		{Method: "DELETE", Path: "/sysjob"},
		{Method: "GET", Path: "/job/remove/{id}"},
		{Method: "GET", Path: "/job/start/{id}"},
		{Method: "GET", Path: "/job-log"},
		{Method: "GET", Path: "/job-log/{id}"},
		{Method: "POST", Path: "/job-log"},
		{Method: "PUT", Path: "/job-log/{id}"},
		{Method: "DELETE", Path: "/job-log"},
	}
	for _, resource := range []string{
		"/account",
		"/account-details",
		"/account-unit",
		"/account-app-role",
		"/account-app-blacklist",
		"/account-change-log",
		"/units",
		"/groups",
		"/containers",
		"/applications",
		"/group-app-blacklist",
		"/pass-ruler",
		"/sms",
		"/cas-login-log",
		"/oauth-log",
		"/oauth-token",
	} {
		routes = append(routes,
			legacyRouteSpec{Method: "GET", Path: resource},
			legacyRouteSpec{Method: "GET", Path: resource + "/{id}"},
			legacyRouteSpec{Method: "POST", Path: resource},
			legacyRouteSpec{Method: "PUT", Path: resource + "/{id}"},
			legacyRouteSpec{Method: "DELETE", Path: resource},
		)
	}
	routes = append(routes,
		legacyRouteSpec{Method: "POST", Path: "/account/unlockPassword"},
		legacyRouteSpec{Method: "POST", Path: "/account/updatePassword"},
		legacyRouteSpec{Method: "POST", Path: "/account/import"},
		legacyRouteSpec{Method: "POST", Path: "/account/importCheck"},
		legacyRouteSpec{Method: "POST", Path: "/account/updatePasswordGetUser"},
		legacyRouteSpec{Method: "POST", Path: "/account/updatePasswordBySelfPhone"},
		legacyRouteSpec{Method: "POST", Path: "/account/updatePasswordBySelf"},
	)
	return routes
}
