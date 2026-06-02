// This file verifies the old uidentity/admin route inventory published by the
// source plugin compatibility layer.

package backend

import "testing"

func TestLegacyRouteSpecIncludesOldAdminContracts(t *testing.T) {
	routes := legacyRouteSpecSet()
	for _, item := range []legacyRouteSpec{
		{Method: "POST", Path: "/account/import"},
		{Method: "POST", Path: "/account/importCheck"},
		{Method: "POST", Path: "/account/updatePassword"},
		{Method: "GET", Path: "/sysjob"},
		{Method: "GET", Path: "/sysjob/{id}"},
		{Method: "GET", Path: "/job/start/{id}"},
		{Method: "GET", Path: "/job/remove/{id}"},
		{Method: "GET", Path: "/health"},
		{Method: "GET", Path: "/metrics"},
		{Method: "GET", Path: "/account"},
		{Method: "GET", Path: "/account/{id}"},
		{Method: "POST", Path: "/account"},
		{Method: "PUT", Path: "/account/{id}"},
		{Method: "DELETE", Path: "/account"},
		{Method: "POST", Path: "/cas/login"},
		{Method: "POST", Path: "/cas/loginByPhone"},
		{Method: "GET", Path: "/cas/proxyValidate"},
		{Method: "GET", Path: "/cas-login/index"},
		{Method: "ALL", Path: "/sso/serviceValidate"},
		{Method: "ALL", Path: "/sso/proxyValidate"},
		{Method: "ALL", Path: "/sso/logout"},
		{Method: "POST", Path: "/ssologin/getToken"},
		{Method: "GET", Path: "/wechat/login"},
		{Method: "GET", Path: "/wechat/loginCallback"},
		{Method: "ALL", Path: "/wechat/callback"},
		{Method: "ALL", Path: "/MP_verify_5osfGmdqMLsyyzYp.txt"},
		{Method: "ALL", Path: "/EcEOCIhE9w.txt"},
		{Method: "POST", Path: "/token/get"},
		{Method: "GET", Path: "/token/getUserInfoByToken"},
		{Method: "POST", Path: "/activate/baseInfo"},
		{Method: "POST", Path: "/user/changePassword"},
		{Method: "GET", Path: "/user/accountAppRole"},
		{Method: "POST", Path: "/oauth/token"},
		{Method: "GET", Path: "/config/cas"},
		{Method: "GET", Path: "/stat/get"},
		{Method: "GET", Path: "/job-log"},
		{Method: "GET", Path: "/job-log/{id}"},
		{Method: "POST", Path: "/job-log"},
		{Method: "PUT", Path: "/job-log/{id}"},
		{Method: "DELETE", Path: "/job-log"},
		{Method: "GET", Path: "/server-monitor"},
		{Method: "POST", Path: "/public/uploadFile"},
	} {
		key := item.Method + " " + item.Path
		if _, ok := routes[key]; !ok {
			t.Fatalf("legacy route missing: %s", key)
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

func legacyRouteSpecsForTest(routes []legacyRouteSpec) map[string]struct{} {
	got := make(map[string]struct{}, len(routes))
	for _, route := range routes {
		got[route.Method+" "+route.Path] = struct{}{}
	}
	return got
}
