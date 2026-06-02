package backend

import (
	"reflect"
	"strings"
	"testing"

	"github.com/gogf/gf/v2/util/gmeta"

	v1 "lina-plugin-linapro-uidentity-cas/backend/api/uidentity/v1"
)

func TestLegacyAPIDTOPathsMatchOldAdminRoutes(t *testing.T) {
	oldRoutes := legacyFullRouteSpecsForTest(oldAdminRouterInventoryForTest())
	for _, contract := range legacyAPIDTOContractsForTest() {
		path := gmeta.Get(contract.req, "path").String()
		method := strings.ToUpper(gmeta.Get(contract.req, "method").String())
		if path == "" || method == "" {
			t.Fatalf("%s is missing g.Meta path or method", contract.name)
		}
		if strings.Contains(path, "/uidentity") {
			t.Fatalf("%s path still uses LinaPro/uidentity namespace: %s", contract.name, path)
		}
		if _, ok := oldRoutes[method+" "+path]; !ok {
			t.Fatalf("%s declares non-legacy route %s %s", contract.name, method, path)
		}
	}
}

func TestLegacyAPIDTOFieldsMatchOldAdminParams(t *testing.T) {
	for _, item := range []struct {
		typ   reflect.Type
		field string
		json  string
	}{
		{reflect.TypeOf(v1.CasPasswordLoginReq{}), "ClientId", "appid"},
		{reflect.TypeOf(v1.CasPasswordLoginReq{}), "Code", "code"},
		{reflect.TypeOf(v1.CasPasswordLoginReq{}), "UUID", "uuid"},
		{reflect.TypeOf(v1.CasPhoneLoginReq{}), "ClientId", "appid"},
		{reflect.TypeOf(v1.CasUnionIDLoginReq{}), "UnionId", "uuid"},
		{reflect.TypeOf(v1.RuntimeTokenIssueReq{}), "ClientId", "appid"},
		{reflect.TypeOf(v1.RuntimeTokenInfoReq{}), "AccessToken", "AccessToken"},
		{reflect.TypeOf(v1.WechatLoginQRReq{}), "ClientId", "appid"},
		{reflect.TypeOf(v1.WechatLoginCallbackReq{}), "ClientId", "appid"},
		{reflect.TypeOf(v1.ActivationStartReq{}), "UUID", "uuid"},
		{reflect.TypeOf(v1.ActivationFaceReq{}), "ChallengeId", "uuid"},
		{reflect.TypeOf(v1.UserUnionIDLookupReq{}), "UnionId", "union_id"},
		{reflect.TypeOf(v1.UserUnionIDBindReq{}), "ChallengeId", "uuid"},
		{reflect.TypeOf(v1.UserUnionIDBindReq{}), "BindType", "bind_type"},
		{reflect.TypeOf(v1.UserPasswordChangeReq{}), "NewPassword", "new_password"},
		{reflect.TypeOf(v1.UserLoginLogsReq{}), "PageNum", "pageIndex"},
		{reflect.TypeOf(v1.ResourceListReq{}), "PageNum", "pageIndex"},
		{reflect.TypeOf(v1.ResourceCreateReq{}), "UnitId", "groupId"},
		{reflect.TypeOf(v1.ResourceCreateReq{}), "GroupIDs", "groupIds"},
		{reflect.TypeOf(v1.ResourceUpdateReq{}), "UnitId", "unitId"},
		{reflect.TypeOf(v1.ResourceDeleteReq{}), "Ids", "ids"},
		{reflect.TypeOf(v1.AccountDetailsGetReq{}), "AccountId", "accountId"},
		{reflect.TypeOf(v1.AccountDetailsDeleteReq{}), "AccountIds", "account_ids"},
		{reflect.TypeOf(v1.AccountUnitCreateReq{}), "AccountId", "accountId"},
		{reflect.TypeOf(v1.AccountUnitCreateReq{}), "UnitId", "unitId"},
		{reflect.TypeOf(v1.LegacyAccountDetailsUpdateBody{}), "Face", "face"},
		{reflect.TypeOf(v1.LegacyAccountAppRoleBody{}), "EmpoweredAccountId", "empoweredAccountId"},
		{reflect.TypeOf(v1.LegacyAccountAppBlacklistCreateBody{}), "Number", "number"},
		{reflect.TypeOf(v1.LegacyApplicationsBody{}), "CallbackUrl", "callbackUrl"},
		{reflect.TypeOf(v1.LegacyPassRulerBody{}), "IntervalStatus", "intervalStatus"},
		{reflect.TypeOf(v1.LegacyOAuthTokenBody{}), "ExpiredAt", "expiredAt"},
		{reflect.TypeOf(v1.OAuthAccessTokenReq{}), "GrantType", "grant_type"},
		{reflect.TypeOf(v1.OAuthAccessTokenReq{}), "ClientSecret", "client_secret"},
		{reflect.TypeOf(v1.LegacyUploadFile{}), "FullPath", "full_path"},
	} {
		field, ok := item.typ.FieldByName(item.field)
		if !ok {
			t.Fatalf("%s.%s is missing", item.typ.Name(), item.field)
		}
		if got := strings.Split(field.Tag.Get("json"), ",")[0]; got != item.json {
			t.Fatalf("%s.%s json tag = %q, want %q", item.typ.Name(), item.field, got, item.json)
		}
	}
}

func legacyAPIDTOContractsForTest() []struct {
	name string
	req  any
} {
	return []struct {
		name string
		req  any
	}{
		{"AccountImportCheckReq", v1.AccountImportCheckReq{}},
		{"AccountImportReq", v1.AccountImportReq{}},
		{"AccountPasswordReq", v1.AccountPasswordReq{}},
		{"AccountPasswordUnlockReq", v1.AccountPasswordUnlockReq{}},
		{"AccountPasswordChallengeReq", v1.AccountPasswordChallengeReq{}},
		{"AccountPasswordPhoneVerifyReq", v1.AccountPasswordPhoneVerifyReq{}},
		{"AccountPasswordSelfResetReq", v1.AccountPasswordSelfResetReq{}},
		{"ActivationStartReq", v1.ActivationStartReq{}},
		{"ActivationFaceReq", v1.ActivationFaceReq{}},
		{"ActivationPasswordReq", v1.ActivationPasswordReq{}},
		{"ActivationPhoneReq", v1.ActivationPhoneReq{}},
		{"ActivationWechatReq", v1.ActivationWechatReq{}},
		{"ActivationWechatStateCreateReq", v1.ActivationWechatStateCreateReq{}},
		{"ActivationWechatCallbackReq", v1.ActivationWechatCallbackReq{}},
		{"ActivationStateReq", v1.ActivationStateReq{}},
		{"CasLoginReq", v1.CasLoginReq{}},
		{"CasPasswordLoginReq", v1.CasPasswordLoginReq{}},
		{"CasPhoneLoginReq", v1.CasPhoneLoginReq{}},
		{"CasUnionIDLoginReq", v1.CasUnionIDLoginReq{}},
		{"CasServiceTicketReq", v1.CasServiceTicketReq{}},
		{"CasServiceValidateReq", v1.CasServiceValidateReq{}},
		{"CasTicketLogoutReq", v1.CasTicketLogoutReq{}},
		{"LegacyCASServiceValidateXMLReq", v1.LegacyCASServiceValidateXMLReq{}},
		{"LegacyCASConfigReq", v1.LegacyCASConfigReq{}},
		{"LegacyLDAPConfigReq", v1.LegacyLDAPConfigReq{}},
		{"LegacyOAuthConfigReq", v1.LegacyOAuthConfigReq{}},
		{"LegacyTokenConfigReq", v1.LegacyTokenConfigReq{}},
		{"LegacyUploadReq", v1.LegacyUploadReq{}},
		{"LegacyHealthReq", v1.LegacyHealthReq{}},
		{"LegacyServerMonitorReq", v1.LegacyServerMonitorReq{}},
		{"LegacyLogSnapshotReq", v1.LegacyLogSnapshotReq{}},
		{"LegacyExternalActionReq", v1.LegacyExternalActionReq{}},
		{"OAuthAuthorizationCodeReq", v1.OAuthAuthorizationCodeReq{}},
		{"OAuthAccessTokenReq", v1.OAuthAccessTokenReq{}},
		{"OAuthAccessTokenInfoReq", v1.OAuthAccessTokenInfoReq{}},
		{"ResourceCreateReq", v1.ResourceCreateReq{}},
		{"ResourceDeleteReq", v1.ResourceDeleteReq{}},
		{"ResourceGetReq", v1.ResourceGetReq{}},
		{"ResourceListReq", v1.ResourceListReq{}},
		{"ResourceUpdateReq", v1.ResourceUpdateReq{}},
		{"AccountDetailsListReq", v1.AccountDetailsListReq{}},
		{"AccountDetailsGetReq", v1.AccountDetailsGetReq{}},
		{"AccountDetailsCreateReq", v1.AccountDetailsCreateReq{}},
		{"AccountDetailsUpdateReq", v1.AccountDetailsUpdateReq{}},
		{"AccountDetailsDeleteReq", v1.AccountDetailsDeleteReq{}},
		{"AccountUnitListReq", v1.AccountUnitListReq{}},
		{"AccountUnitGetReq", v1.AccountUnitGetReq{}},
		{"AccountUnitCreateReq", v1.AccountUnitCreateReq{}},
		{"AccountUnitUpdateReq", v1.AccountUnitUpdateReq{}},
		{"AccountUnitDeleteReq", v1.AccountUnitDeleteReq{}},
		{"AccountAppRoleListReq", v1.AccountAppRoleListReq{}},
		{"AccountAppRoleGetReq", v1.AccountAppRoleGetReq{}},
		{"AccountAppRoleCreateReq", v1.AccountAppRoleCreateReq{}},
		{"AccountAppRoleUpdateReq", v1.AccountAppRoleUpdateReq{}},
		{"AccountAppRoleDeleteReq", v1.AccountAppRoleDeleteReq{}},
		{"AccountAppBlacklistListReq", v1.AccountAppBlacklistListReq{}},
		{"AccountAppBlacklistGetReq", v1.AccountAppBlacklistGetReq{}},
		{"AccountAppBlacklistCreateReq", v1.AccountAppBlacklistCreateReq{}},
		{"AccountAppBlacklistUpdateReq", v1.AccountAppBlacklistUpdateReq{}},
		{"AccountAppBlacklistDeleteReq", v1.AccountAppBlacklistDeleteReq{}},
		{"AccountChangeLogListReq", v1.AccountChangeLogListReq{}},
		{"AccountChangeLogGetReq", v1.AccountChangeLogGetReq{}},
		{"AccountChangeLogCreateReq", v1.AccountChangeLogCreateReq{}},
		{"AccountChangeLogUpdateReq", v1.AccountChangeLogUpdateReq{}},
		{"AccountChangeLogDeleteReq", v1.AccountChangeLogDeleteReq{}},
		{"UnitsListReq", v1.UnitsListReq{}},
		{"UnitsGetReq", v1.UnitsGetReq{}},
		{"UnitsCreateReq", v1.UnitsCreateReq{}},
		{"UnitsUpdateReq", v1.UnitsUpdateReq{}},
		{"UnitsDeleteReq", v1.UnitsDeleteReq{}},
		{"GroupsListReq", v1.GroupsListReq{}},
		{"GroupsGetReq", v1.GroupsGetReq{}},
		{"GroupsCreateReq", v1.GroupsCreateReq{}},
		{"GroupsUpdateReq", v1.GroupsUpdateReq{}},
		{"GroupsDeleteReq", v1.GroupsDeleteReq{}},
		{"ContainersListReq", v1.ContainersListReq{}},
		{"ContainersGetReq", v1.ContainersGetReq{}},
		{"ContainersCreateReq", v1.ContainersCreateReq{}},
		{"ContainersUpdateReq", v1.ContainersUpdateReq{}},
		{"ContainersDeleteReq", v1.ContainersDeleteReq{}},
		{"ApplicationsListReq", v1.ApplicationsListReq{}},
		{"ApplicationsGetReq", v1.ApplicationsGetReq{}},
		{"ApplicationsCreateReq", v1.ApplicationsCreateReq{}},
		{"ApplicationsUpdateReq", v1.ApplicationsUpdateReq{}},
		{"ApplicationsDeleteReq", v1.ApplicationsDeleteReq{}},
		{"GroupAppBlacklistListReq", v1.GroupAppBlacklistListReq{}},
		{"GroupAppBlacklistGetReq", v1.GroupAppBlacklistGetReq{}},
		{"GroupAppBlacklistCreateReq", v1.GroupAppBlacklistCreateReq{}},
		{"GroupAppBlacklistUpdateReq", v1.GroupAppBlacklistUpdateReq{}},
		{"GroupAppBlacklistDeleteReq", v1.GroupAppBlacklistDeleteReq{}},
		{"PassRulerListReq", v1.PassRulerListReq{}},
		{"PassRulerGetReq", v1.PassRulerGetReq{}},
		{"PassRulerCreateReq", v1.PassRulerCreateReq{}},
		{"PassRulerUpdateReq", v1.PassRulerUpdateReq{}},
		{"PassRulerDeleteReq", v1.PassRulerDeleteReq{}},
		{"SmsListReq", v1.SmsListReq{}},
		{"SmsGetReq", v1.SmsGetReq{}},
		{"SmsCreateReq", v1.SmsCreateReq{}},
		{"SmsUpdateReq", v1.SmsUpdateReq{}},
		{"SmsDeleteReq", v1.SmsDeleteReq{}},
		{"CasLoginLogListReq", v1.CasLoginLogListReq{}},
		{"CasLoginLogGetReq", v1.CasLoginLogGetReq{}},
		{"CasLoginLogCreateReq", v1.CasLoginLogCreateReq{}},
		{"CasLoginLogUpdateReq", v1.CasLoginLogUpdateReq{}},
		{"CasLoginLogDeleteReq", v1.CasLoginLogDeleteReq{}},
		{"OAuthLogListReq", v1.OAuthLogListReq{}},
		{"OAuthLogGetReq", v1.OAuthLogGetReq{}},
		{"OAuthLogCreateReq", v1.OAuthLogCreateReq{}},
		{"OAuthLogUpdateReq", v1.OAuthLogUpdateReq{}},
		{"OAuthLogDeleteReq", v1.OAuthLogDeleteReq{}},
		{"OAuthTokenListReq", v1.OAuthTokenListReq{}},
		{"OAuthTokenGetReq", v1.OAuthTokenGetReq{}},
		{"OAuthTokenCreateReq", v1.OAuthTokenCreateReq{}},
		{"OAuthTokenUpdateReq", v1.OAuthTokenUpdateReq{}},
		{"OAuthTokenDeleteReq", v1.OAuthTokenDeleteReq{}},
		{"SmsSendReq", v1.SmsSendReq{}},
		{"StatsReq", v1.StatsReq{}},
		{"RuntimeTokenIssueReq", v1.RuntimeTokenIssueReq{}},
		{"RuntimeTokenInfoReq", v1.RuntimeTokenInfoReq{}},
		{"UserUnionIDLookupReq", v1.UserUnionIDLookupReq{}},
		{"UserUnionIDBindReq", v1.UserUnionIDBindReq{}},
		{"UserPasswordChangeReq", v1.UserPasswordChangeReq{}},
		{"UserPhoneChangeReq", v1.UserPhoneChangeReq{}},
		{"UserEmailChangeReq", v1.UserEmailChangeReq{}},
		{"UserQQChangeReq", v1.UserQQChangeReq{}},
		{"UserWechatUnbindReq", v1.UserWechatUnbindReq{}},
		{"UserWechatRebindStateCreateReq", v1.UserWechatRebindStateCreateReq{}},
		{"UserWechatRebindCallbackReq", v1.UserWechatRebindCallbackReq{}},
		{"UserWechatRebindStateReq", v1.UserWechatRebindStateReq{}},
		{"UserInfoReq", v1.UserInfoReq{}},
		{"UserLoginLogsReq", v1.UserLoginLogsReq{}},
		{"UserApplicationsReq", v1.UserApplicationsReq{}},
		{"UserAppRolesReq", v1.UserAppRolesReq{}},
		{"UserAppRoleCreateReq", v1.UserAppRoleCreateReq{}},
		{"UserAppRoleUpdateReq", v1.UserAppRoleUpdateReq{}},
		{"WechatLoginQRReq", v1.WechatLoginQRReq{}},
		{"WechatLoginCallbackReq", v1.WechatLoginCallbackReq{}},
		{"WechatLoginQRResultReq", v1.WechatLoginQRResultReq{}},
	}
}

func legacyFullRouteSpecsForTest(routes []legacyRouteSpec) map[string]struct{} {
	got := make(map[string]struct{}, len(routes))
	for _, route := range routes {
		path := route.Path
		if strings.HasPrefix(path, "/api/v1/") || strings.HasPrefix(path, "/sso/") || strings.HasPrefix(path, "/ssologin/") {
			got[strings.ToUpper(route.Method)+" "+path] = struct{}{}
			continue
		}
		if strings.HasPrefix(path, "/wechat/callback") || strings.HasPrefix(path, "/MP_verify_") || strings.HasPrefix(path, "/EcEOCI") || path == "/" || path == "/info" {
			got[strings.ToUpper(route.Method)+" "+path] = struct{}{}
			continue
		}
		got[strings.ToUpper(route.Method)+" /api/v1"+path] = struct{}{}
	}
	return got
}
