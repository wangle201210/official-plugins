// This file registers legacy uidentity/admin root routes for compatibility
// clients. The routes reuse the plugin controller and host middlewares while
// leaving LinaPro core framework contracts unchanged.

package backend

import (
	"strings"

	"lina-core/pkg/plugin/pluginhost"
	uidentitycontroller "lina-plugin-linapro-uidentity-cas/backend/internal/controller/uidentity"
)

type legacyResourceRoute struct {
	Path     string
	Resource string
}

type legacyRouteSpec struct {
	Method string
	Path   string
}

var legacyResourceRoutes = []legacyResourceRoute{
	{Path: "/account", Resource: "accounts"},
	{Path: "/account-details", Resource: "account-details"},
	{Path: "/account-unit", Resource: "account-unit"},
	{Path: "/account-app-role", Resource: "account-app-role"},
	{Path: "/account-app-blacklist", Resource: "account-app-blacklist"},
	{Path: "/account-change-log", Resource: "account-change-log"},
	{Path: "/units", Resource: "units"},
	{Path: "/groups", Resource: "groups"},
	{Path: "/containers", Resource: "containers"},
	{Path: "/applications", Resource: "applications"},
	{Path: "/group-app-blacklist", Resource: "group-app-blacklist"},
	{Path: "/pass-ruler", Resource: "pass-ruler"},
	{Path: "/sms", Resource: "sms"},
	{Path: "/cas-login-log", Resource: "cas-login-logs-legacy"},
	{Path: "/oauth-log", Resource: "oauth-log"},
	{Path: "/oauth-token", Resource: "oauth-token"},
}

var legacyPublicRoutes = []legacyRouteSpec{
	{Method: "POST", Path: "/account/updatePasswordGetUser"},
	{Method: "POST", Path: "/account/updatePasswordBySelfPhone"},
	{Method: "POST", Path: "/account/updatePasswordBySelf"},
	{Method: "GET", Path: "/health"},
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
	{Method: "ALL", Path: "/sso/serviceValidate"},
	{Method: "ALL", Path: "/sso/proxyValidate"},
	{Method: "GET", Path: "/sso/login"},
	{Method: "ALL", Path: "/sso/logout"},
	{Method: "POST", Path: "/ssologin/getToken"},
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
	{Method: "POST", Path: "/user/getUserInfo"},
	{Method: "GET", Path: "/user/getUserCasLoginLog"},
	{Method: "GET", Path: "/user/accountAppList"},
	{Method: "GET", Path: "/user/accountAppRole"},
	{Method: "POST", Path: "/user/accountAppRole"},
	{Method: "POST", Path: "/user/accountAppRoleUpdate"},
	{Method: "POST", Path: "/user/changeWechatQr"},
	{Method: "POST", Path: "/user/changeWechatState"},
	{Method: "POST", Path: "/user/changeWechatCallBack"},
	{Method: "GET", Path: "/oauth/login"},
	{Method: "POST", Path: "/oauth/login"},
	{Method: "ALL", Path: "/oauth/auth"},
	{Method: "ALL", Path: "/oauth/authorize"},
	{Method: "POST", Path: "/oauth/token"},
	{Method: "ALL", Path: "/oauth/test"},
	{Method: "POST", Path: "/sms/send"},
	{Method: "GET", Path: "/stat/get"},
}

var legacyProtectedRoutes = []legacyRouteSpec{
	{Method: "GET", Path: "/sysjob"},
	{Method: "GET", Path: "/sysjob/{id}"},
	{Method: "POST", Path: "/sysjob"},
	{Method: "PUT", Path: "/sysjob"},
	{Method: "DELETE", Path: "/sysjob"},
	{Method: "GET", Path: "/job-log"},
	{Method: "GET", Path: "/job-log/{id}"},
	{Method: "POST", Path: "/job-log"},
	{Method: "PUT", Path: "/job-log/{id}"},
	{Method: "DELETE", Path: "/job-log"},
	{Method: "POST", Path: "/account/unlockPassword"},
	{Method: "POST", Path: "/account/updatePassword"},
	{Method: "POST", Path: "/account/import"},
	{Method: "POST", Path: "/account/importCheck"},
	{Method: "GET", Path: "/job/start/{id}"},
	{Method: "GET", Path: "/job/remove/{id}"},
	{Method: "GET", Path: "/config/cas"},
	{Method: "GET", Path: "/config/ldap"},
	{Method: "GET", Path: "/config/oauth"},
	{Method: "GET", Path: "/config/token"},
	{Method: "GET", Path: "/server-monitor"},
	{Method: "POST", Path: "/public/uploadFile"},
	{Method: "GET", Path: "/log/watch"},
	{Method: "POST", Path: "/ldap/sync"},
	{Method: "POST", Path: "/job/start"},
	{Method: "POST", Path: "/job/remove"},
}

var legacyRootSSORoutes = []legacyRouteSpec{
	{Method: "ALL", Path: "/sso/serviceValidate"},
	{Method: "ALL", Path: "/sso/proxyValidate"},
	{Method: "GET", Path: "/sso/login"},
	{Method: "ALL", Path: "/sso/logout"},
	{Method: "POST", Path: "/ssologin/getToken"},
}

var legacyRootWechatRoutes = []legacyRouteSpec{
	{Method: "ALL", Path: "/wechat/callback"},
	{Method: "ALL", Path: "/MP_verify_5osfGmdqMLsyyzYp.txt"},
	{Method: "ALL", Path: "/EcEOCIhE9w.txt"},
}

func registerLegacyRoutes(routes pluginhost.RouteRegistrar, middlewares pluginhost.RouteMiddlewares, legacyController *uidentitycontroller.LegacyController) {
	routes.Group("/api/v1", func(group pluginhost.RouteGroup) {
		registerLegacyBaseMiddlewares(group, middlewares)
		registerLegacyPublicRoutes(group, legacyController)
		group.Group("/", func(group pluginhost.RouteGroup) {
			group.Middleware(
				middlewares.Auth(),
				middlewares.Tenancy(),
				middlewares.Permission(),
			)
			registerLegacyProtectedRoutes(group, legacyController)
			registerLegacyResourceRoutes(group, legacyController)
		})
	})
	routes.Group("/sso", func(group pluginhost.RouteGroup) {
		registerLegacyBaseMiddlewares(group, middlewares)
		group.ALL("/serviceValidate", legacyController.CasServiceValidateXML)
		group.ALL("/proxyValidate", legacyController.CasServiceValidateXML)
		group.GET("/login", legacyController.SSOLogin)
		group.ALL("/logout", legacyController.SSOLogout)
	})
	routes.Group("/ssologin", func(group pluginhost.RouteGroup) {
		registerLegacyBaseMiddlewares(group, middlewares)
		group.POST("/getToken", legacyController.SSOLoginToken)
	})
	routes.Group("/", func(group pluginhost.RouteGroup) {
		registerLegacyBaseMiddlewares(group, middlewares)
		group.ALL("/wechat/callback", legacyController.WechatCallback)
		group.ALL("/MP_verify_5osfGmdqMLsyyzYp.txt", legacyController.WechatVerifyMP)
		group.ALL("/EcEOCIhE9w.txt", legacyController.WechatVerifyEc)
	})
}

func registerLegacyBaseMiddlewares(group pluginhost.RouteGroup, middlewares pluginhost.RouteMiddlewares) {
	group.Middleware(
		middlewares.NeverDoneCtx(),
		middlewares.HandlerResponse(),
		middlewares.CORS(),
		middlewares.RequestBodyLimit(),
		middlewares.Ctx(),
	)
}

func registerLegacyPublicRoutes(group pluginhost.RouteGroup, legacyController *uidentitycontroller.LegacyController) {
	group.POST("/account/updatePasswordGetUser", legacyController.AccountPasswordChallenge)
	group.POST("/account/updatePasswordBySelfPhone", legacyController.AccountPasswordPhoneVerify)
	group.POST("/account/updatePasswordBySelf", legacyController.AccountPasswordSelfReset)
	group.GET("/health", legacyController.Health)
	group.GET("/cas/login", legacyController.CasLoginByCookie)
	group.POST("/cas/login", legacyController.CasPasswordLogin)
	group.POST("/cas/loginByPhone", legacyController.CasPhoneLogin)
	group.DELETE("/cas/tickets/{ticket}", legacyController.CasTicketLogout)
	group.GET("/cas/proxyValidate", legacyController.CasServiceValidateXML)
	group.POST("/cas/getLoginQr", legacyController.CasGetLoginQR)
	group.GET("/cas/loginByQr", legacyController.CasLoginByQR)
	group.POST("/cas/getCasLoginQrRes", legacyController.CasGetLoginQRResult)
	group.GET("/cas/loginByUnionID", legacyController.CasUnionIDLogin)
	group.GET("/cas-login/index", legacyController.CasLoginIndex)
	group.ALL("/sso/serviceValidate", legacyController.CasServiceValidateXML)
	group.ALL("/sso/proxyValidate", legacyController.CasServiceValidateXML)
	group.GET("/sso/login", legacyController.SSOLogin)
	group.ALL("/sso/logout", legacyController.SSOLogout)
	group.POST("/ssologin/getToken", legacyController.SSOLoginToken)
	group.POST("/token/get", legacyController.RuntimeTokenIssue)
	group.GET("/token/getUserInfoByToken", legacyController.RuntimeTokenInfo)
	group.GET("/wechat/login", legacyController.WechatLogin)
	group.GET("/wechat/loginCallback", legacyController.WechatLoginCallback)
	group.POST("/activate/baseInfo", legacyController.ActivationStart)
	group.POST("/activate/face", legacyController.ActivationFace)
	group.POST("/activate/password", legacyController.ActivationPassword)
	group.POST("/activate/phone", legacyController.ActivationPhone)
	group.POST("/activate/wechatQr", legacyController.ActivationWechatQR)
	group.GET("/activate/wechatScan", legacyController.ActivationWechatScan)
	group.POST("/activate/state", legacyController.ActivationState)
	group.POST("/user/getByUnionID", legacyController.UserGetByUnionID)
	group.GET("/user/bindUnionIDCallBack", legacyController.UserBindUnionIDCallback)
	group.POST("/user/bindUnionID", legacyController.UserBindUnionID)
	group.POST("/user/changePassword", legacyController.UserChangePassword)
	group.POST("/user/changePhone", legacyController.UserChangePhone)
	group.POST("/user/changeEmail", legacyController.UserChangeEmail)
	group.POST("/user/changeQQ", legacyController.UserChangeQQ)
	group.POST("/user/unbindWechat", legacyController.UserUnbindWechat)
	group.POST("/user/getUserInfo", legacyController.UserInfo)
	group.GET("/user/getUserCasLoginLog", legacyController.UserLoginLogs)
	group.GET("/user/accountAppList", legacyController.UserApplications)
	group.GET("/user/accountAppRole", legacyController.UserAppRoles)
	group.POST("/user/accountAppRole", legacyController.UserAppRoleCreate)
	group.POST("/user/accountAppRoleUpdate", legacyController.UserAppRoleUpdate)
	group.POST("/user/changeWechatQr", legacyController.UserChangeWechatQR)
	group.POST("/user/changeWechatState", legacyController.UserChangeWechatState)
	group.POST("/user/changeWechatCallBack", legacyController.UserWechatRebindCallback)
	group.GET("/oauth/login", legacyController.OAuthLogin)
	group.POST("/oauth/login", legacyController.OAuthLogin)
	group.ALL("/oauth/auth", legacyController.OAuthAuthorize)
	group.ALL("/oauth/authorize", legacyController.OAuthAuthorize)
	group.POST("/oauth/token", legacyController.OAuthToken)
	group.ALL("/oauth/test", legacyController.OAuthTest)
	group.POST("/sms/send", legacyController.SmsSend)
	group.GET("/stat/get", legacyController.Stats)
}

func registerLegacyProtectedRoutes(group pluginhost.RouteGroup, legacyController *uidentitycontroller.LegacyController) {
	group.GET("/sysjob", legacyController.SysJobList)
	group.GET("/sysjob/{id}", legacyController.SysJobGet)
	group.POST("/sysjob", legacyController.SysJobExternalAction("sysjob-create"))
	group.PUT("/sysjob", legacyController.SysJobExternalAction("sysjob-update"))
	group.DELETE("/sysjob", legacyController.SysJobExternalAction("sysjob-delete"))
	group.GET("/job-log", legacyController.JobLogList)
	group.GET("/job-log/{id}", legacyController.JobLogGet)
	group.POST("/job-log", legacyController.ExternalAction("job-log-create"))
	group.PUT("/job-log/{id}", legacyController.ExternalAction("job-log-update"))
	group.DELETE("/job-log", legacyController.ExternalAction("job-log-delete"))
	group.POST("/account/unlockPassword", legacyController.AccountPasswordUnlock)
	group.POST("/account/updatePassword", legacyController.AccountPasswordUpdate)
	group.POST("/account/import", legacyController.AccountImport)
	group.POST("/account/importCheck", legacyController.AccountImportCheck)
	group.GET("/job/start/{id}", legacyController.SysJobExternalAction("job-start"))
	group.GET("/job/remove/{id}", legacyController.SysJobExternalAction("job-remove"))
	group.GET("/config/cas", legacyController.LegacyCASConfig)
	group.GET("/config/ldap", legacyController.LegacyLDAPConfig)
	group.GET("/config/oauth", legacyController.LegacyOAuthConfig)
	group.GET("/config/token", legacyController.LegacyTokenConfig)
	group.GET("/server-monitor", legacyController.ServerMonitor)
	group.POST("/public/uploadFile", legacyController.Upload)
	group.GET("/log/watch", legacyController.LogSnapshot)
	group.POST("/ldap/sync", legacyController.ExternalAction("ldap-sync"))
	group.POST("/job/start", legacyController.ExternalAction("job-start"))
	group.POST("/job/remove", legacyController.ExternalAction("job-remove"))
}

func registerLegacyResourceRoutes(group pluginhost.RouteGroup, legacyController *uidentitycontroller.LegacyController) {
	for _, route := range legacyResourceRoutes {
		group.GET(route.Path, legacyController.ResourceList(route.Resource))
		group.GET(route.Path+"/{id}", legacyController.ResourceGet(route.Resource))
		group.POST(route.Path, legacyController.ResourceCreate(route.Resource))
		group.PUT(route.Path+"/{id}", legacyController.ResourceUpdate(route.Resource))
		group.DELETE(route.Path, legacyController.ResourceDelete(route.Resource))
	}
}

func allLegacyRouteSpecs() []legacyRouteSpec {
	result := make([]legacyRouteSpec, 0, len(legacyPublicRoutes)+len(legacyProtectedRoutes)+len(legacyRootSSORoutes)+len(legacyRootWechatRoutes)+len(legacyResourceRoutes)*5)
	result = append(result, legacyPublicRoutes...)
	result = append(result, legacyProtectedRoutes...)
	result = append(result, legacyRootSSORoutes...)
	result = append(result, legacyRootWechatRoutes...)
	for _, route := range legacyResourceRoutes {
		result = append(result,
			legacyRouteSpec{Method: "GET", Path: route.Path},
			legacyRouteSpec{Method: "GET", Path: route.Path + "/{id}"},
			legacyRouteSpec{Method: "POST", Path: route.Path},
			legacyRouteSpec{Method: "PUT", Path: route.Path + "/{id}"},
			legacyRouteSpec{Method: "DELETE", Path: route.Path},
		)
	}
	return result
}

func legacyRouteSpecSet() map[string]struct{} {
	result := make(map[string]struct{})
	for _, route := range allLegacyRouteSpecs() {
		result[strings.ToUpper(route.Method)+" "+route.Path] = struct{}{}
	}
	return result
}
