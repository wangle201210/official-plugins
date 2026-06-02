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
		{Method: "GET", Path: "/account"},
		{Method: "GET", Path: "/account/{id}"},
		{Method: "POST", Path: "/account"},
		{Method: "PUT", Path: "/account/{id}"},
		{Method: "DELETE", Path: "/account"},
		{Method: "POST", Path: "/cas/login"},
		{Method: "POST", Path: "/cas/loginByPhone"},
		{Method: "GET", Path: "/cas/proxyValidate"},
		{Method: "GET", Path: "/sso/serviceValidate"},
		{Method: "POST", Path: "/token/get"},
		{Method: "GET", Path: "/token/getUserInfoByToken"},
		{Method: "POST", Path: "/activate/baseInfo"},
		{Method: "POST", Path: "/user/changePassword"},
		{Method: "GET", Path: "/user/accountAppRole"},
		{Method: "POST", Path: "/oauth/token"},
		{Method: "GET", Path: "/config/cas"},
		{Method: "GET", Path: "/stat/get"},
		{Method: "GET", Path: "/server-monitor"},
		{Method: "POST", Path: "/public/uploadFile"},
	} {
		key := item.Method + " " + item.Path
		if _, ok := routes[key]; !ok {
			t.Fatalf("legacy route missing: %s", key)
		}
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
